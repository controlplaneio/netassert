const test = require('ava')
const { findLocalPortsToTest, replaceNegationOperator, isNegation, stripProtocol } = require('./ports')

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

test('replaceNegationOperator removes negation operator', t => {
  let port = '-80'

  let replaced = replaceNegationOperator(port)

  t.is(replaced, '80', 'expected negation operator to be reomved')
})

test('replaceNegationOperator only removes at start of string', t => {
  let port = 'TCP:-80'

  let replaced = replaceNegationOperator(port)

  t.is(replaced, 'TCP:-80', 'expected negation operator to be reomved')
})

test('isNegation return true if first char is -', t => {
  let port = "-80"
  t.true(isNegation(port))
})

test('isNegation return false if first char is not -', t => {
  let port = '80'
  t.false(isNegation(port))
})

test('stripProtocol strips negated protocol', t => {
  let portSpec = '-TCP:80'
  t.is(stripProtocol(portSpec), '-80')
})

test('stripProtocol strips protocol', t => {
  let portSpec = 'UDP:22'
  t.is(stripProtocol(portSpec), '22')
})

test('stripProtocol returns unchanged with no protocol', t => {
  let portSpec = '-80'
  t.is(stripProtocol(portSpec), '-80')
})
