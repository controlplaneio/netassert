const negationOperator = '-'

module.exports = {
  NEGATION_OPERATOR: negationOperator,
  replaceNegationOperator (port) {
    const regex = new RegExp(`^${negationOperator}`, 'g')
    return port.replace(regex, '')
  }
}

