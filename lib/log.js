function log (...args) {
  if (process.env.DEBUG) {
    console.log(...args)
  }
}

module.exports = { log }
