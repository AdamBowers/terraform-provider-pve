package diagnostic

type severity string

const (
	SEVERITY_Error   severity = "Error"
	SEVERITY_Warning severity = "Warning"
	SEVERITY_Info    severity = "Info"
	SEVERITY_Verbose severity = "Verbose"
)
