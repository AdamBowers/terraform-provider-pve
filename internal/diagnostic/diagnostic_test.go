package diagnostic

import (
	"strings"
	"testing"
)

type mockCustomDiagnostic struct {
	sev   severity
	sum   string
	det   string
	trace string
}

func (m mockCustomDiagnostic) Summary() string           { return m.sum }
func (m mockCustomDiagnostic) Detail() string            { return m.det }
func (m mockCustomDiagnostic) Severity() severity        { return m.sev }
func (m mockCustomDiagnostic) Stacktrace() string        { return m.trace }
func (m mockCustomDiagnostic) String() string            { return m.sum }
func (m mockCustomDiagnostic) Equal(to IDiagnostic) bool { return false }

func TestDiagnosticInterfaceGetters(t *testing.T) {
	d := &diagnostic{
		severity:   SEVERITY_Error,
		summary:    "SumMsg",
		detail:     "DetMsg",
		category:   Category_InvalidParamter,
		stacktrace: Stack{Frame{Func: "Worker"}},
	}

	if d.Summary() != "SumMsg" {
		t.Errorf("Summary() mismatch: %q", d.Summary())
	}
	if d.Detail() != "DetMsg" {
		t.Errorf("Detail() mismatch: %q", d.Detail())
	}
	if d.Severity() != SEVERITY_Error {
		t.Errorf("Severity() mismatch: %q", d.Severity())
	}
	if !strings.Contains(d.Stacktrace(), "Worker") {
		t.Errorf("Stacktrace() processing missing string segments: %q", d.Stacktrace())
	}
}

func TestDiagnosticStringLayouts(t *testing.T) {
	d := &diagnostic{
		severity: SEVERITY_Warning,
		category: Category_InvalidParamter,
		summary:  "Bad Setup",
		detail:   "Check configs",
	}

	strOutput := d.String()
	expectedFields := []string{"Severity: Warning", "Category: Invalid Parameter", "Summary:\n\tBad Setup", "Detail:\n\tCheck configs"}

	for _, field := range expectedFields {
		if !strings.Contains(strOutput, field) {
			t.Errorf("diagnostic.String() string output dropped tracking segment: %q", field)
		}
	}
}

func TestDiagnosticEqualLogic(t *testing.T) {
	base := &diagnostic{severity: SEVERITY_Error, summary: "A", detail: "B", category: "C"}

	if base.Equal(nil) {
		t.Error("Equal(nil) should return false")
	}

	// Dynamic API Interface Verification Fallback Branch
	mockExternal := mockCustomDiagnostic{sev: SEVERITY_Error, sum: "A", det: "B", trace: ""}
	if !base.Equal(mockExternal) {
		t.Error("Interface evaluation verification branch matching fallback rules failed")
	}
}

func TestFactoryConstructors(t *testing.T) {
	sum, det := "S", "D"

	errDiag := Error(sum, det).(*diagnostic)
	if errDiag.severity != SEVERITY_Error || errDiag.stacktrace == nil {
		t.Error("Error() constructor failed to configure defaults")
	}

	warnDiag := Warning(sum, det).(*diagnostic)
	if warnDiag.severity != SEVERITY_Warning {
		t.Error("Warning() mismatch")
	}

	infoDiag := Info(sum, det).(*diagnostic)
	if infoDiag.severity != SEVERITY_Info {
		t.Error("Info() mismatch")
	}

	verbDiag := Verbose(sum, det).(*diagnostic)
	if verbDiag.severity != SEVERITY_Verbose {
		t.Error("Verbose() mismatch")
	}
}
