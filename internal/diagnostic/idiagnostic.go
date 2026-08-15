package diagnostic

import "fmt"

type IDiagnostic interface {
	fmt.Stringer

	// Returns a summary message of the diagnostic
	Summary() string

	// Returns a detailed message of the diagnostic
	Detail() string

	// Returns the severity of the diagnostic
	Severity() Severity

	// Checks if this diagnostic is equal to the provided diagnostic
	Equal(to IDiagnostic) bool
}
