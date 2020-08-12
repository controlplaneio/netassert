const nmap = require('node-nmap')

const { log } = require('./log')

const DISABLE_NMAP_DISCOVERY = '-Pn '
const SCAN_UDP_WITH_VERSION = '-sU -sV '
const SCAN_TIMEOUT_MILLIS = 30000

// Converts a list of port specs (e.g. [ "-TCP:80", TCP:443" ]) and returns the commandline flags for use with nmap
function portSpecsToNmapOptions(ports, protocol) {
  return `-p ${ports.map((x) => `${protocol.toUpperCase()[0]}:${x}`).join(',')} `
}

// Builds a string of commandline options to pass to nmap
function buildNmapOptions(ports, protocol, rng = Math.random) {
  let optionString = ''
  if (protocol.toLowerCase() == 'udp') {
    optionString += SCAN_UDP_WITH_VERSION
  }

  optionString += DISABLE_NMAP_DISCOVERY
  optionString += portSpecsToNmapOptions(ports, protocol)

  const jitter = (3 * rng()).toFixed(3)
  const scanTimeoutSeconds = SCAN_TIMEOUT_MILLIS / 1000
  //
  // -T1 with 10s initial-rtt-timeout and jitter
  optionString += '--initial-rtt-timeout 10s --max-retries 100 --max-rate 1 '
  optionString += `--scan-delay ${jitter} --host-timeout ${scanTimeoutSeconds}s`

  return optionString
}

function scan(host, ports, protocol, done) {
  nmap.nmapLocation = '/usr/bin/nmap' // We run in a container so we can be sure of this
  const nmapOptions = buildNmapOptions(ports, protocol)

  log(`query string: ${nmapOptions}, for ${protocol}`)

  let quickscan = new nmap.NmapScan(host, nmapOptions)

  quickscan.scanTimeout = SCAN_TIMEOUT_MILLIS
  quickscan.on('complete', scanResults => {
    log(`results for ${protocol}`)
    log(JSON.stringify(scanResults, null, 2))
    log(`proto ${protocol}:`, JSON.stringify(ports, null, 2))

    return done(null, parseResults(scanResults))
  })
  quickscan.on('error', error => done(error))

  quickscan.startScan()
}

// Takes scanResults that nmap calls back with and returns a list of open ports
function parseResults(scanResults) {
  // (rem): non of these guard clauses should ever trigger
  if (scanResults.length == 0) {
    throw new Error('scan results was empty - expected one result')
  }

  if (scanResults.length > 1) {
    throw new Error('scan results had more than one resultset - expected one result')
  }

  if (scanResults[0].openPorts === undefined) {
    throw new Error('scan results had no openPorts property')
  }

  return scanResults[0].openPorts.map((openPort) =>  parseInt(openPort.port, 10))
}


module.exports = {
  buildNmapOptions,
  portSpecsToNmapOptions,
  parseResults,
  scan,
  SCAN_TIMEOUT_MILLIS
}
