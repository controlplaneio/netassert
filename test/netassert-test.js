import test from 'ava'

const nmap = require('node-nmap')
nmap.nmapLocation = '/usr/bin/nmap' // default
const yaml = require('yaml-js')
const fs = require('fs')

const debug = (process.env.DEBUG === '0' ? false : (!!process.env.DEBUG ? true : !!process.env.REMOTE_DEBUG))
process.setMaxListeners(200)

// console.log(`# DEBUG: ${debug} - ENV: ${process.env.DEBUG}`)

function log () {
  if (debug) {
    // process.stdout.write('# ')
    console.log.apply(null, arguments)
  }
}

var filename = 'test.yaml'
let contents = fs.readFileSync(filename, 'utf8')
var tests = yaml.load(contents)

const negationOperator = '-'

log(tests)

const runTests = (tests) => {
  Object.keys(tests).forEach((testType) => {

    switch (testType) {
      case 'k8s':
      case 'kubernetes':
        runKubernetesTests(tests[testType])
        break

      case 'host':
      case 'instance':
        runHostTests(tests[testType])
        break
      default:
        console.error(`Unknown test type ${testType}`)
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
      case 'localhost':
        runHostLocalTests(tests[testType])
        break

      default:
        if (testType.substr(0, 1) === '_') {
          runHostLocalTests(tests[testType])
        } else {
          console.error(`Unknown test type ${testType}`)
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
      console.error(`Invalid spec, empty port(s) found [${portsToTest.join(',')}]`)
      process.exit(1)
    }

    log('all ports to test', portsToTest)
    log(`test: ${host}, ports ${portsToTest.join(',')}`, JSON.stringify(portsToTest))

    const tcpPortsToTest = portsToTest.filter(tcpOnly)
    tcpPortsToTest.forEach(port => test.cb(openTcp, host, [port]))

    const udpPortsToTest = portsToTest.filter(udpOnly)
    udpPortsToTest.forEach(port => test.cb(openUdp, host, [port]))

    const httpPortsToTest = portsToTest.filter(httpOnly)
    httpPortsToTest.forEach(port => test.cb(openHttp, host, [port]))

    log(`complete: ${host}, ports ${portsToTest.join(',')}`, JSON.stringify(portsToTest))

    log()
  })
  log()
}

// ---

function openTcp (t, host, portsToTest) {
  assertPortsOpen(t, host, portsToTest, 'tcp')
}

function openUdp (t, host, portsToTest) {
  assertPortsOpen(t, host, portsToTest, 'udp')
}

function openHttp (t, host, portsToTest) {
  assertPortsOpen(t, host, portsToTest, 'http')
}

openTcp.title = (providedTitle, host, expectedPorts) => {
  expectedPorts = getTestName(expectedPorts)
  return `${providedTitle} ${host} TCP:${expectedPorts.join(',')}`.trim()
}

openUdp.title = (providedTitle, host, expectedPorts) => {
  expectedPorts = getTestName(expectedPorts)
  return `${providedTitle} ${host} UDP:${expectedPorts.join(',')}`.trim()
}

openHttp.title = (providedTitle, host, expectedPorts) => {
  expectedPorts = getTestName(expectedPorts)
  return `${providedTitle} ${host} HTTP:${expectedPorts.join(',')}`.trim()
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

  log('assertPortsOpen start', protocol)

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

  let nmapPortsArgument = portsToTest.map((x) => `${protocol.toUpperCase()[0]}:${x}`).join(',')
  // set HTTP(s) queries to TCP
  nmapPortsArgument = nmapPortsArgument.replace(/H:/, 'T:')

  let nmapQueryString = `-Pn -p ${nmapPortsArgument}`

  switch (protocol) {
    case 'tcp':
      nmapQueryString = `${nmapQueryString} -sT` // TCP Connect() scan
      break
    case 'udp':
      nmapQueryString = `-sU -sV ${nmapQueryString}`
      break
    case 'http':
      nmapQueryString = `${nmapQueryString} -sT` // TCP Connect() scan
      nmapQueryString = `${nmapQueryString} --script=http-get --script-args http-get.path=/,http-get.showResponse` // HTTP GET
      break
    case 'https':
      nmapQueryString = `${nmapQueryString} -sT` // TCP Connect() scan
      nmapQueryString = `${nmapQueryString} --script=http-get --script-args http-get.path=/,http-get.showResponse,http-get.forceTls` // HTTP GET
      break
    default:
      throw new Error(`Protocol ${protocol} not supported`)
  }

  const jitter = (3 * Math.random()).toFixed(3)
  const scanTimeout = 30000
  // -T1 with 10s initial-rtt-timeout and jitter
  nmapQueryString = `${nmapQueryString} --initial-rtt-timeout 10s --max-retries 100 --max-rate 1`
  nmapQueryString = `${nmapQueryString} --scan-delay ${jitter} --host-timeout ${scanTimeout / 1000}s`

  let host = hosts[0]

  log('query string: ', nmapQueryString, `for ${protocol}`)

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
        // bypass TCP/UDP protocol checking as HTTP is checked via TCP
        if (protocol !== 'http' && openPort.protocol !== protocol) {
          log('protocol fail', protocol, 'vs', openPort.protocol)
          t.fail(`protocol mismatch: ${openPort.protocol} != ${protocol}`)
        }

        log(`open port on ${host}`, openPort)
        let isExtendedValidationPass = true
        if (protocol === 'http') {
          // TODO(ajm) validate status code and allow different paths
          if (openPort.scriptOutput !== '\n  GET / -> 200 OK\n') {
            isExtendedValidationPass = false
          }
        }

        if (isExtendedValidationPass) {
          foundPorts.push(parseInt(openPort.port, 10))
        } else {
          log(`additional checks failed for ${protocol} on port ${openPort} for ${host}`)
        }
      })
    }

    expectedPorts.forEach(expectedPort => {
      log('all ports, this one', expectedPort, 'protocol', protocol)
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

function tcpOnly (ports) {
  ports = replaceNegationOperator(ports)
  if (ports.substr(0, 4) === 'TCP:') {
    return true
  }
  let portsProtocol = ports.split(':')[0]
  return !['UDP', 'ICMP', 'HTTP'].includes(portsProtocol)
}

function udpOnly (ports) {
  return (replaceNegationOperator(ports).substr(0, 4) === 'UDP:')
}

function httpOnly (ports) {
  return (replaceNegationOperator(ports).substr(0, 5) === 'HTTP:')
}

// TODO(ajm) not implemented
function icmpOnly (ports) {
  return (replaceNegationOperator(ports).substr(0, 5) === 'ICMP:')
}

function replaceNegationOperator (port) {
  const regex = new RegExp(`^${negationOperator}`, 'g')
  return port.replace(regex, '')
}

runTests(tests)
