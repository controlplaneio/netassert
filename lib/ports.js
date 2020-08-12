const negationOperator = '-'

function splitPortsString (ports) {
  return ports.toString().split(/[\s,]+/)
}

function isNegation (port) {
  return port.substr(0, 1) === negationOperator
}

function replaceNegationOperator (port) {
  const regex = new RegExp(`^${negationOperator}`, 'g')
  return port.replace(regex, '')
}

module.exports = {
  splitPortsString,
  isNegation,
  replaceNegationOperator,
  findLocalPortsToTest (hostSpec) {
    const isSingleHost = !Array.isArray(hostSpec)
    const portsToTest = isSingleHost ? splitPortsString(hostSpec) : hostSpec.map(port => splitPortsString(port)).flat()

    const emptyPorts = portsToTest.filter(x => x === '')
    if (emptyPorts.length > 0) {
      console.error(`Invalid spec, empty port(s) found [${portsToTest.join(',')}]`)
      process.exit(1)
    }

    return portsToTest
  },
  stripProtocol (port) {
    const closed = isNegation(port)
    port = replaceNegationOperator(port).split(':')
    return (closed ? negationOperator : '') + port[port.length - 1]
  }
}
