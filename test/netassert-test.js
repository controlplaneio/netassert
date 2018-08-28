import test from 'ava'

const nmap = require('node-nmap')
nmap.nmapLocation = '/usr/bin/nmap' // default
const yaml = require('yaml-js');
const fs = require('fs');

const debug = !!process.env.DEBUG
process.setMaxListeners(200)

function log() {
  if (debug) {
    console.log.apply(null, arguments)
  }
}

var filename = "test.yaml"
let contents = fs.readFileSync(filename, 'utf8');
var tests = yaml.load(contents)

const negationOperator = '-'

log(tests)

const runTests = (tests) => {
  Object.keys(tests).forEach((testType) => {

    switch (testType) {
      case "k8s":
      case "kubernetes":
        runKubernetesTests(tests[testType])
        break;

      case "host":
      case "instance":
        runHostTests(tests[testType])
        break;
      default:
        console.error(`Unknown test type ${testType}`);
        process.exit(1)
    }

    log()
  })
}

const runKubernetesTests = (tests) => {
  log('k8s tests')

}

const runHostTests = (tests) => {
  log('host tests', tests)

  Object.keys(tests).forEach((testType) => {

    switch (testType) {
      case "localhost":
        runHostLocalTests(tests[testType])
        break;

      default:
        if (testType.substr(0, 1) === '_') {
          runHostLocalTests(tests[testType])
        } else {
          console.error(`Unknown test type ${testType}`);
          process.exit(1)
        }
    }

    log()
  })
}

const splitPortsString = (ports) => {
  return ports.toString().split(/[\s,]+/)
}

const runHostLocalTests = (tests) => {
  log('host local tests', tests)

  Object.keys(tests).forEach((host) => {
    var portsToTest = tests[host]
    if (!Array.isArray(portsToTest)) {
      portsToTest = splitPortsString(portsToTest)
    } else {
      portsToTest = portsToTest.map(port => splitPortsString(port))
      portsToTest = [].concat.apply([], portsToTest)
    }

    const zeroLengthPorts = portsToTest.filter(x => x == '')
    if (zeroLengthPorts.length > 0) {
      console.error(`Invalid spec, empty port(s) found [${portsToTest.join(",")}]`);
      process.exit(1)
    }

    log('all ports to test', portsToTest)
    log(`test ${host}, ports ${portsToTest.join(",")}`, JSON.stringify(portsToTest))

    const tcpPortsToTest = portsToTest.filter(tcpOnly)
    tcpPortsToTest.forEach(port => test.cb(openTcp, host, [port]))

    const udpPortsToTest = portsToTest.filter(udpOnly)
    udpPortsToTest.forEach(port => test.cb(openUdp, host, [port]))

    log()
  })
  log()
}

// ---

function openTcp(t, host, portsToTest) {
  assertPortsOpen(t, host, portsToTest, 'tcp')
}

function openUdp(t, host, portsToTest) {
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
    const isNegation = (port.substr(0, 1) === negationOperator)
    port = replaceNegationOperator(port).split(':')
    return `${port[port.length - 1]} ${isNegation ? 'closed' : 'open'}`
  })
}

// ---

const assertPortsOpen = (t, hosts, portsToTest, protocol = 'tcp') => {

  if (!Array.isArray(hosts)) {
    hosts = hosts.split(' ')
  }

  if (!Array.isArray(portsToTest)) {
    portsToTest = [portsToTest]
  }
  if (portsToTest.length < 1) {
    return t.end()
  }

  let expectedPorts = portsToTest.map(port => {
    const isNegation = (port.substr(0, 1) === negationOperator)
    port = replaceNegationOperator(port).split(':')
    return (isNegation ? negationOperator : '') + port[port.length - 1]
  })

  portsToTest = portsToTest.map(port => {
    port = port.split(':')
    return replaceNegationOperator(port[port.length - 1])
  })
  log('ports to test', portsToTest, portsToTest.length)
  log('expected ports', expectedPorts, expectedPorts.length)

  let expectedTests = hosts.length * portsToTest.length
  log('expected', expectedTests)
  t.plan(expectedTests)

  let nmapQueryString = `-Pn -p ${portsToTest.map((x) => `${protocol.toUpperCase()[0]}:${x}`).join(',')}`

  if (protocol == 'udp') {
    nmapQueryString = `-sU -sV ${nmapQueryString}`
  }

  const jitter = (3 * Math.random()).toFixed(3)
  const scanTimeout = 30000
  // -T1 with 10s initial-rtt-timeout and jitter
  nmapQueryString = `${nmapQueryString} --initial-rtt-timeout 10s --max-retries 100 --max-rate 1`
  nmapQueryString = `${nmapQueryString} --scan-delay ${jitter} --host-timeout ${scanTimeout/1000}s`

  let host = hosts[0]

  log("query string: ", nmapQueryString, `for ${protocol}`)

  let quickscan = new nmap.NmapScan(host, `${nmapQueryString}`)
  quickscan.scanTimeout = scanTimeout
  quickscan.on('complete', function (scanResults) {

    log(`results for ${protocol}`)
    log(JSON.stringify(scanResults, null, 2))
    log(`proto ${protocol}:`, JSON.stringify(portsToTest, null, 2))

    let foundPorts = []
    if (scanResults.length && scanResults[0].openPorts) {
      if (scanResults.length > 1) {
        t.fail(`Only one host supported per scan, found ${scanResults.length}`)
      }

      scanResults[0].openPorts.forEach((openPort) => {
        if (openPort.protocol != protocol) {
          t.fail(`protocol mismatch: ${openPort.protocol} != ${protocol}`)
        }

        log(`open port on ${host}`, openPort)
        foundPorts.push(parseInt(openPort.port, 10))
      })
    }

    expectedPorts.forEach(expectedPort => {
      log('all ports, this one', expectedPort)
      const isNegation = (expectedPort.substr(0, 1) === negationOperator)
      if (isNegation) {
        expectedPort = parseInt(expectedPort.substr(1), 10)
        log('asserting', expectedPort, 'is NOT IN', foundPorts)
        t.falsy(
          foundPorts.includes(expectedPort),
          `${host}: expected ${protocol}:${expectedPort} to be closed, found [${foundPorts.join(',')}]`
        )

      } else {
        log('asserting', expectedPort, 'in', foundPorts)
        expectedPort = parseInt(expectedPort, 10)
        t.truthy(
          foundPorts.includes(expectedPort),
          `${host}: expected ${protocol}:${expectedPort} to be open, found [${foundPorts.join(',')}]`
        )
      }
    })

    log('done')
    t.end()
  })

  quickscan.on('error', function (error) {
    log(error)
    t.fail(error)
    t.end()
  })

  quickscan.startScan()
}

// ---

function tcpOnly(ports) {
  ports = replaceNegationOperator(ports)
  if (ports.substr(0, 4) === 'TCP:') {
    return true
  }
  return ports.substr(0, 4) !== 'UDP:' && ports.substr(0, 5) !== 'ICMP:'
}

function udpOnly(ports) {
  return (replaceNegationOperator(ports).substr(0, 4) === 'UDP:')
}

// TODO(ajm) not implemented
function icmpOnly(ports) {
  return (replaceNegationOperator(ports).substr(0, 5) === 'ICMP:')
}

function replaceNegationOperator(port) {
  const regex = new RegExp(`^${negationOperator}`, "g")
  return port.replace(regex, '')
}

runTests(tests)
