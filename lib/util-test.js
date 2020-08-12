const test = require('ava')
const { replaceNegationOperator } = require('./util')

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
