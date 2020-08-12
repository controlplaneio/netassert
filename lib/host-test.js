const test = require('ava')
const { findLocalPortsToTest } = require('./host')

test('findLocalPortsToTest deals with comma separated string', t => {
  const portSpec = "80, 443"
  const portsToTest = findLocalPortsToTest(portSpec)

  t.deepEqual(portsToTest, [ "80", "443" ])
})

test('findLocalPortsToTest splits and flattens array of strings', t => {
  const portSpec = [ "80, 443", "22", "UDP:1234" ]
  const portsToTest = findLocalPortsToTest(portSpec)

  t.deepEqual(portsToTest, [ "80", "443", "22", "UDP:1234" ])
})
