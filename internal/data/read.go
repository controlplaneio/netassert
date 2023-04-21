package data

import (
	"fmt"
	"os"
	"path/filepath"
)

// List of file extensions we support
const (
	fileExtensionYAML = `.yaml`
	fileExtensionYML  = `.yml`
)

// ReadTestsFromDir - Reads tests cases from .yaml and .yml file present in a directory
// does not recursively read files
func ReadTestsFromDir(path string) (Tests, error) {
	if path == "" {
		return nil, fmt.Errorf("input dir parameter cannot be empty string")
	}

	var testCases Tests

	fp, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open dir containing tests %q: %w", path, err)
	}

	// we do not recursively read all the YAML and YML files
	// the depth is only 1 level
	files, err := fp.ReadDir(0)
	if err != nil {
		return nil, fmt.Errorf("unable to read contents of the directory%q: %w", path, err)
	}

	for _, f := range files {

		ext := filepath.Ext(f.Name())

		if ext != fileExtensionYAML && ext != fileExtensionYML {
			continue
		}

		tcFile := filepath.Join(path, f.Name())

		tc, err := ReadTestsFromFile(tcFile)
		if err != nil {
			return nil, err
		}

		testCases = append(testCases, tc...)

		// this is a multi-files validation each time new tests are added
		if err := testCases.Validate(); err != nil {
			return nil, fmt.Errorf("validation of tests from file %q failed: %w", tcFile, err)
		}

	}

	return testCases, nil
}

// ReadTestsFromFile - reads tests from a file containing a list of Test
func ReadTestsFromFile(fileName string) (Tests, error) {
	if fileName == "" {
		return nil, fmt.Errorf("input fileName parameter can not be empty string")
	}

	if fileName == "-" {
		return NewFromReader(os.Stdin)
	}

	fp, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to open file %q containing tests: %w", fileName, err)
	}

	defer func() {
		closeErr := fp.Close()
		if closeErr != nil {
			err = fmt.Errorf("unable to close file: %q, %w", fileName, closeErr)
		}
	}()

	return NewFromReader(fp)
}
