package logger

import (
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/go-hclog"
)

// NewHCLogger - return an instance of the logger
func NewHCLogger(logLevel, appName string, w io.Writer) hclog.Logger {
	hcLevel := hclog.LevelFromString(strings.ToUpper(logLevel))

	if hcLevel == hclog.NoLevel {
		hcLevel = hclog.Info
	}

	var includeLocation bool

	if hcLevel <= hclog.Debug {
		// check if hcLevel is DEBUG or more verbose
		includeLocation = true
	}

	l := hclog.New(&hclog.LoggerOptions{
		Name:            fmt.Sprintf("[%s]", appName),
		Level:           hcLevel,
		Output:          w,
		JSONFormat:      false,
		IncludeLocation: includeLocation,
	})

	return l
}
