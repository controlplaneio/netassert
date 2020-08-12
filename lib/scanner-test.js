const test = require('ava')
const { buildNmapOptions, portSpecsToNmapOptions } = require('./scanner')


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
  const ports = [80,443,22]
  const protocol = 'TCP'
  t.is(portSpecsToNmapOptions(ports, protocol), '-p T:80,T:443,T:22 ')
})

test('portSpecsToNmapOptions with many ports (lowercase)', t => {
  const ports = [80,443,22]
  const protocol = 'tcp'
  t.is(portSpecsToNmapOptions(ports, protocol), '-p T:80,T:443,T:22 ')
})

test('buildNmapOptions for non UDP', t => {
  const ports = [80]
  const protocol = 'TCP'
  const rng = () => 1;
  const nmapOptions = buildNmapOptions(ports, protocol, rng)

  t.is(nmapOptions, '-Pn -p T:80 --initial-rtt-timeout 10s --max-retries 100 --max-rate 1 --scan-delay 3.000 --host-timeout 30s')
})

test('buildNmapOptions for UDP', t => {
  const ports = [80]
  const protocol = 'UDP'
  const rng = () => 1;
  const nmapOptions = buildNmapOptions(ports, protocol, rng)

  t.is(nmapOptions, '-sU -sV -Pn -p U:80 --initial-rtt-timeout 10s --max-retries 100 --max-rate 1 --scan-delay 3.000 --host-timeout 30s')
})
