const debug = (process.env.DEBUG === '0' ? false : (process.env.DEBUG ? true : !!process.env.REMOTE_DEBUG))
process.setMaxListeners(200)

function log () {
  if (debug) {
    // process.stdout.write('# ')
    console.log.apply(null, arguments)
  }
}

module.exports = { log }
