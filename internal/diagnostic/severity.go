package diagnostic

type Severity string

const (
	SEVERITY_Error   Severity = "Error"
	SEVERITY_Warning Severity = "Warning"
	SEVERITY_Info    Severity = "Info"
	SEVERITY_Verbose Severity = "Verbose"
)
