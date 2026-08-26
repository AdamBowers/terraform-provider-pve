package diagnostic

import "fmt"

type IDiagnostic interface {
	fmt.Stringer

	// Returns a summary message of the diagnostic
	Summary() string

	// Returns a detailed message of the diagnostic
	Detail() string

	// Returns the stacktrace of the diagnostic
	Stacktrace() string

	// Returns the severity of the diagnostic
	Severity() severity

	// Checks if this diagnostic is equal to the provided diagnostic
	Equal(to IDiagnostic) bool
}
