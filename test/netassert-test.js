const test = require('ava')
const nmap = require('node-nmap')
nmap.nmapLocation = '/usr/bin/nmap' // default
const { join } = require('path')

const { isNegation, replaceNegationOperator, findLocalPortsToTest, stripProtocol } = require('../lib/ports')
const { scan } = require('../lib/scanner')
const { loadTests } = require('../lib/io')
const { log } = require('../lib/log')

// we really shouldnt need to do this the warning is only triggered when you add >10 event handlers for the same event.
// TODO(rem): we should diagnose the problem in node-nmap
// TODO(rem): we are on a very old version of node-nmap this probem may have been fixed in a more recent version
process.setMaxListeners(200)

const tests = loadTests(join(__dirname, 'test.yaml')) // should be user suppliable

const runTestsFromManifest = (tests) => {
  log('Starting reading test manifest')
  // (rem) This script is only ever run with known yaml input which will be massaged to only include
  // a manifest that requires it to nmap ports to other hosts.  This fn is legacy
  Object.keys(tests).forEach((testType) => {
    switch (testType) {
      case 'host':
      case 'instance': // (rem) why instance?! this doesn't seem to be ever a type supplied
        runTests(tests[testType])
        break
      default: // should be unreachable
        console.error(`Unknown test type ${testType}`)
        process.exit(1)
    }
  })
}

const runTests = (tests) => {
  log('Running tests', tests)

  Object.keys(tests).forEach((testType) => {
    switch (testType) {
      case 'localhost':
        runHostLocalTests(tests[testType])
        break

      default:
        // (rem): this is important refer to the format that is produced by the driver bash script
        if (testType.substr(0, 1) === '_') {
          runHostLocalTests(tests[testType])
        } else {
          console.error(`Unknown test type ${testType}`)
          process.exit(1)
        }
    }
  })
}

const runHostLocalTests = (tests) => {
  log('local tests', tests)

  Object.keys(tests).forEach((host) => {
    var portsToTest = findLocalPortsToTest(tests[host])

    log('all ports to test', portsToTest)
    log(`test ${host}, ports ${portsToTest.join(',')}`, JSON.stringify(portsToTest))

    const tcpPortsToTest = portsToTest.filter(tcpOnly)
    tcpPortsToTest.forEach(port => test.cb(openTcp, host, [port]))

    const udpPortsToTest = portsToTest.filter(udpOnly)
    udpPortsToTest.forEach(port => test.cb(openUdp, host, [port]))
  })
}

// ---

function openTcp (t, host, portsToTest) {
  assertPortsOpen(t, host, portsToTest, 'tcp')
}

function openUdp (t, host, portsToTest) {
  assertPortsOpen(t, host, portsToTest, 'udp')
}

openTcp.title = (providedTitle, host, expectedPorts) => {
  expectedPorts = getTestName(expectedPorts)
  return `${providedTitle} ${host} TCP:${expectedPorts.join(',')}`.trim()
}

openUdp.title = (providedTitle, host, expectedPorts) => {
  expectedPorts = getTestName(expectedPorts)
  return `${providedTitle} ${host} UDP:${expectedPorts.join(',')}`.trim()
}

const getTestName = (expectedPorts) => {
  return expectedPorts.map(port => {
    const closed = isNegation(port)
    port = replaceNegationOperator(port).split(':')
    return `${port[port.length - 1]} ${closed ? 'closed' : 'open'}`
  })
}

// ---

const assertPortsOpen = (t, hosts, ports, protocol = 'tcp') => {
  if (!Array.isArray(hosts)) {
    hosts = hosts.split(' ')
  }

  if (ports.length < 1) {
    return t.end()
  }

  // Remove the protocols preserving the negation operator
  // [ "-TCP:80", "TCP:443" ] -> [ "-80", "443" ]
  const portExpectations = ports.map(stripProtocol)

  // Extract just the port numbers - we'll loop over the results and compre with portExpectations to check whether they
  // are open or closed
  // [ "-80", "443" ] -> [ "80", "443" ]
  const portsToTest = portExpectations.map(replaceNegationOperator)

  log('ports to test', portsToTest, portsToTest.length)
  log('expected ports', portExpectations, portExpectations.length)

  // (rem): this is computation is redundant we only ever scan one host in this function
  const testCount = hosts.length * portsToTest.length

  log('expected', testCount)
  t.plan(testCount)

  // TODO(rem): we're only scanning the first host here. This function is only ever called with a single host
  const host = hosts[0]
  scan(host, portsToTest, protocol, (error, foundPorts) => {
    if (error) {
      log(error)
      t.fail(error)
      return t.end()
    }

    portExpectations.forEach(portExpectation => {
      log(`all ports, this one ${portExpectation}`)
      const port = parseInt(portExpectation.substr(1), 10)
      const closed = isNegation(portExpectation)

      if (closed) {
        log(`asserting ${port}, is NOT IN, ${foundPorts}`)
        t.falsy(
          foundPorts.includes(port),
          `${host}: expected ${protocol}:${port} to be closed, found [${foundPorts.join(',')}]`
        )
      } else {
        log('asserting', portExpectation, 'in', foundPorts)
        portExpectation = parseInt(portExpectation, 10)
        t.truthy(
          foundPorts.includes(portExpectation),
          `${host}: expected ${protocol}:${portExpectation} to be open, found [${foundPorts.join(',')}]`
        )
      }
    })

    log('done')
    t.end()
  })
}

// ---

function tcpOnly (ports) {
  ports = replaceNegationOperator(ports)
  if (ports.substr(0, 4) === 'TCP:') {
    return true
  }
  return ports.substr(0, 4) !== 'UDP:' && ports.substr(0, 5) !== 'ICMP:'
}

function udpOnly (ports) {
  return (replaceNegationOperator(ports).substr(0, 4) === 'UDP:')
}

runTestsFromManifest(tests)
