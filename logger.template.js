const logIngestorURL = 'https://your-log-ingestor-endpoint.com/ingest';
const LEVELS = { silly: 0, debug: 10, info: 20, warn: 30, error: 40 };
const DEFAULT_LEVEL = info
const BUFFER_SIZE = 50; // quantity of logs to buffer before sending
const logBuffer = [];


// Returns current log level (env variable)
function getLogLevel() {
	const env = (DEFAULT_LEVEL).toLowerCase();
	return Object.prototype.hasOwnProperty.call(LEVELS, env) ? env : DEFAULT_LEVEL;
}


// Checks if message should be logged given current level
function shouldLog(level) {
	const current = getLogLevel();
	return LEVELS[level] >= LEVELS[current];
}


// Safe stringify for meta object (handles circular references)
function safeStringify(obj) {
	try {
		return JSON.stringify(obj);
	} catch (e) {
		return String(obj);
	}
}

// 
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

	logBuffer.push(safeStringify(logObj));
	
	if (logBuffer.length >= BUFFER_SIZE) {
		// send log to external service asynchronously
		sendLog(logBuffer.join('\n'));
		logBuffer.length = 0; // clear buffer
	}
	
	// remove comment to debug by printing to console
    // const stream = level === 'error' ? process.stderr : process.stdout;
	// stream.write(line + '\n');
}

// Sends log line to external log ingestor service
async function sendLog(logs) {
	try {
		await fetch(logIngestorURL, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: logs
		});
	}
	catch (error) {}
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
