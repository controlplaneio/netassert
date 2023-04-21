package data

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

// TAPResult - outputs result of tests into a TAP format
func (ts *Tests) TAPResult(w io.Writer) error {
	if ts == nil {
		return fmt.Errorf("empty ts")
	}

	if len(*ts) < 1 {
		return fmt.Errorf("no test were found")
	}

	header := "TAP version 14\n"
	header += fmt.Sprintf("1..%v\n", len(*ts))
	_, err := fmt.Fprint(w, header)
	if err != nil {
		return err
	}

	for index, test := range *ts {
		result := ""
		switch test.Pass {
		case true:
			result = fmt.Sprintf("ok %v - %v", index+1, test.Name)
		case false:
			frEscaped, err := yaml.Marshal(&test.FailureReason) // frEscaped ends with "\n"
			if err != nil {
				return err
			}
			result = fmt.Sprintf("not ok %v - %v", index+1, test.Name)
			result += fmt.Sprintf("\n  ---\n  reason: %s  ...", frEscaped)
		}

		if _, err := fmt.Fprintln(w, result); err != nil {
			return err
		}
	}

	return nil
}
