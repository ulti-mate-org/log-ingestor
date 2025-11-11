const LEVELS = { silly: 0, debug: 10, info: 20, warn: 30, error: 40 };
const DEFAULT_LEVEL = process.env.LOG_LEVEL || 'silly'; // TODO: change to 'info' in production
const LOG_TYPE = process.env.LOG_TYPE || 'json'; // read once
const BUFFER_SIZE = 5; // quantity of logs to buffer before sending
const logBuffer = [];

/**
 * Pads number with leading zero if needed
 */
function pad(n) {
	return String(n).padStart(2, '0');
}

/**
 * Returns current date/time as DD-MM-YYYY - hh-mm-ss
 */
function nowParts(date = new Date()) {
	const DD = pad(date.getDate());
	const MM = pad(date.getMonth() + 1);
	const YYYY = date.getFullYear();

	const hh = pad(date.getHours());
	const mm = pad(date.getMinutes());
	const ss = pad(date.getSeconds());

	return {
		date: `${DD}-${MM}-${YYYY}`,
		time: `${hh}:${mm}:${ss}`,
	};
}

/**
 * Returns current log level (env variable)
 */
function getLogLevel() {
	const env = (DEFAULT_LEVEL).toLowerCase();
	return Object.prototype.hasOwnProperty.call(LEVELS, env) ? env : DEFAULT_LEVEL;
}

/**
 * Checks if message should be logged given current level
 */
function shouldLog(level) {
	const current = getLogLevel();
	return LEVELS[level] >= LEVELS[current];
}

/**
 * Formats log line with timestamp and level
 */
function formatLine(level, message) {
	const { date, time } = nowParts();
	return `[${date}]   [${time}]   [${level.toUpperCase()}]   -   ${message}`;
}

/**
 * Safe stringify for meta object (handles circular references)
 */
function safeStringify(obj) {
	try {
		return JSON.stringify(obj);
	} catch (e) {
		return String(obj);
	}
}

// Writes human-friendly, timestamped lines to stdout/stderr using ESM
// or JSON format based on LOG_TYPE env variable
// Format: [DD-MM-YYYY]   [hh-mm-ss]   [<TYPE>]   -   message
// Levels: debug, info, warn, error (error -> stderr)
function write(level, msg, meta) {
	if (!shouldLog(level)) return;

	const isJSON = LOG_TYPE === 'json';
	const extra = {};
	let message = msg;

	if (msg instanceof Error) {
		message = msg.message;
		extra.stack = msg.stack;
	}

	if (meta && typeof meta === 'object') {
		Object.assign(extra, meta);
	}

	const logObj = {
		timestamp: new Date().toISOString(),
		service: "backend",
		level,
		message,
		...(Object.keys(extra).length ? { meta: extra } : {})
	};

	const line = isJSON
		? safeStringify(logObj)
		: formatLine(level, message) +
		  (Object.keys(extra).length ? ' | ' + safeStringify(extra) : '');

	logBuffer.push(line);
	
	if (logBuffer.length >= BUFFER_SIZE) {
		// send log to external service asynchronously
		sendLog(logBuffer.join('\n'));
		logBuffer.length = 0; // clear buffer
	}
	
	const stream = level === 'error' ? process.stderr : process.stdout;
	stream.write(line + '\n');
}

/**
 * Sends log line to external log ingestor service
 * // send logs to http://log-ingestor:8080/ingest // TODO: use env (config) variables for this
 * // TODO: implement queuing and retry mechanism + buffers and stuff
 */
async function sendLog(logs) {
	try {
		await fetch('http://log-ingestor:8080/ingest', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: logs
		});
	}
	catch (error) {
		// TODO: handle error on submitting log
	}
}

export default {
	silly: (msg, meta) => write('silly', msg, meta),
	debug: (msg, meta) => write('debug', msg, meta),
	info:  (msg, meta) => write('info', msg, meta),
	warn:  (msg, meta) => write('warn', msg, meta),
	error: (msg, meta) => write('error', msg, meta),
	// helper to expose current level
	currentLevel: getLogLevel,
};
