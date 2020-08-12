const nmap = require('node-nmap')

const DISABLE_NMAP_DISCOVERY = '-Pn '
const SCAN_UDP_WITH_VERSION = '-sU -sV '
const SCAN_TIMEOUT_MILLIS = 30000

function portSpecsToNmapOptions(ports, protocol) {
  return `-p ${ports.map((x) => `${protocol.toUpperCase()[0]}:${x}`).join(',')} `
}

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
  nmap.nmapLocation = '/usr/bin/nmap' // default
  const nmapOptions = buildNmapOptions(ports, protocol)

  console.log(`query string: ${nmapOptions}, for ${protocol}`)

  let quickscan = new nmap.NmapScan(host, nmapOptions)

  quickscan.scanTimeout = SCAN_TIMEOUT_MILLIS
  quickscan.on('complete', scanResults => done(null, scanResults))
  quickscan.on('error', error => done(error))

  quickscan.startScan()
}


module.exports = {
  buildNmapOptions,
  portSpecsToNmapOptions,
  scan,
  SCAN_TIMEOUT_MILLIS
}
