const test = require('ava')
const { buildNmapOptions, portSpecsToNmapOptions, parseResults } = require('./scanner')

test('portSpecsToNmapOptions with one port', t => {
  const ports = [1234]
  const protocol = 'UDP'
  t.is(portSpecsToNmapOptions(ports, protocol), '-p U:1234 ')
})

test('portSpecsToNmapOptions with one port (lowercase)', t => {
  const ports = [1234]
  const protocol = 'udp'
  t.is(portSpecsToNmapOptions(ports, protocol), '-p U:1234 ')
})

test('portSpecsToNmapOptions with many ports', t => {
  const ports = [80, 443, 22]
  const protocol = 'TCP'
  t.is(portSpecsToNmapOptions(ports, protocol), '-p T:80,T:443,T:22 ')
})

test('portSpecsToNmapOptions with many ports (lowercase)', t => {
  const ports = [80, 443, 22]
  const protocol = 'tcp'
  t.is(portSpecsToNmapOptions(ports, protocol), '-p T:80,T:443,T:22 ')
})

test('buildNmapOptions for non UDP', t => {
  const ports = [80]
  const protocol = 'TCP'
  const rng = () => 1
  const nmapOptions = buildNmapOptions(ports, protocol, rng)

  t.is(nmapOptions, '-Pn -p T:80 --initial-rtt-timeout 10s --max-retries 100 --max-rate 1 --scan-delay 3.000 --host-timeout 30s')
})

test('buildNmapOptions for UDP', t => {
  const ports = [80]
  const protocol = 'UDP'
  const rng = () => 1
  const nmapOptions = buildNmapOptions(ports, protocol, rng)

  t.is(nmapOptions, '-sU -sV -Pn -p U:80 --initial-rtt-timeout 10s --max-retries 100 --max-rate 1 --scan-delay 3.000 --host-timeout 30s')
})

// Needs integration test harness
test.todo('scan runs nmap correctly')

test('parseResults throws for no results', t => {
  const results = []
  t.throws(() => parseResults(results))
})

test('parseResults throws for more than 1 resultset', t => {
  const results = [{}, {}, {}]
  t.throws(() => parseResults(results))
})

test('parseResults throws for no openPorts property', t => {
  const results = [{}]
  t.throws(() => parseResults(results))
})

test('parseResults maps to list of open ports', t => {
  const results = [{ openPorts: [{ port: '80' }, { port: '443' }] }]

  t.deepEqual(parseResults(results), [80, 443])
})
