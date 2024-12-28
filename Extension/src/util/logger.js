function logInfo(message) {
  console.info(
    "%c[UMN Lecture Downloader Worker INFO] " + `%c${message}`,
    "color: #22d3ee;",
    "color: #ffffff;"
  );
}

function logError(message) {
  console.error(
    "%c[UMN Lecture Downloader Worker ERROR] " + `%c${message}`,
    "color: #f87171;",
    "color: #ffffff;"
  );
}

function logWarning(message) {
  console.warn(
    "%c[UMN Lecture Downloader Worker WARNING] " + `%c${message}`,
    "color: #fb923c;",
    "color: #ffffff;"
  );
}

function logSuccess(message) {
  console.log(
    "%c[UMN Lecture Downloader Worker SUCCESS] " + `%c${message}`,
    "color: #4ade80;",
    "color: #ffffff;"
  );
}
