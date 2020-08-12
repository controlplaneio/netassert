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

module.exports = {
  buildNmapOptions,
  portSpecsToNmapOptions,
  SCAN_TIMOUT_MILLIS
}
