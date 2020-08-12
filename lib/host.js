function splitPortsString (ports) {
  return ports.toString().split(/[\s,]+/)
}

module.exports = {
  splitPortsString,
  findLocalPortsToTest (hostSpec) {
    const isSingleHost = !Array.isArray(hostSpec)
    let portsToTest = isSingleHost ? splitPortsString(hostSpec) : hostSpec.map(port => splitPortsString(port)).flat()

    const emptyPorts = portsToTest.filter(x => x == '')
    if (emptyPorts.length > 0) {
      console.error(`Invalid spec, empty port(s) found [${portsToTest.join(',')}]`)
      process.exit(1)
    }

    return portsToTest
  }
}
