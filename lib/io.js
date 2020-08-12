const fs = require('fs')
const yaml = require('js-yaml')
const { readFileSync } = require('fs')

module.exports = {
  // Loads and parses a test yaml file from the supplied absolute path.
  // Returns an object representing the yaml file
  loadTests (p) {
    const doc = yaml.safeLoad(readFileSync(p, 'utf8'))
    return doc
  }
}
