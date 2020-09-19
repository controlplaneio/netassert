function log (...args) {
  if (process.env.DEBUG !== '0' && process.env.DEBUG !== "") {
    console.log('#', ...args)
  }
}

module.exports = { log }
