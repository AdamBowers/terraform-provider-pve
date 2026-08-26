package diagnostic

import "strings"

type diagnostic struct {
	severity   severity
	summary    string
	detail     string
	category   category
	stacktrace Stack
}

var _ IDiagnostic = (*diagnostic)(nil)

func (d *diagnostic) Summary() string    { return d.summary }
func (d *diagnostic) Detail() string     { return d.detail }
func (d *diagnostic) Severity() severity { return d.severity }
func (d *diagnostic) Stacktrace() string {
	if d.stacktrace == nil {
		return ""
	}
	return d.stacktrace.String()
}
func (d *diagnostic) String() string {
	sb := strings.Builder{}

	sb.WriteString("Severity: ")
	sb.WriteString(string(d.severity))
	sb.WriteByte('\n')

	if string(d.category) != "" {
		sb.WriteString("Category: ")
		sb.WriteString(string(d.category))
		sb.WriteByte('\n')
	}

	if d.summary != "" {
		sb.WriteString("Summary:\n\t")
		sb.WriteString(d.summary)
		sb.WriteByte('\n')
	}

	if d.detail != "" {
		sb.WriteString("Detail:\n\t")
		sb.WriteString(d.detail)
		sb.WriteByte('\n')
	}

	trace := d.stacktrace.String()
	if trace != "" {
		sb.WriteString("Trace:\n")
		sb.WriteString(trace)
	}

	return sb.String()
}
func (d *diagnostic) Equal(to IDiagnostic) bool {
	if to == nil {
		return false
	}
	if concrete, ok := to.(*diagnostic); ok {
		return d.severity == concrete.severity &&
			d.summary == concrete.summary &&
			d.detail == concrete.detail &&
			d.category == concrete.category
	}
	return d.Severity() == to.Severity() &&
		d.Summary() == to.Summary() &&
		d.Detail() == to.Detail() &&
		d.Stacktrace() == to.Stacktrace()
}

// Error creates an Error severity diagnostic.
func Error(summary string, detail string, opts ...Option) IDiagnostic {
	d := &diagnostic{
		severity: SEVERITY_Error,
		summary:  summary,
		detail:   detail,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(d)
		}
	}

	if d.stacktrace == nil {
		d.stacktrace = GenerateStack(3) // 3 is to skip runtime.callers, Generate Stack, this function
	}
	return d
}

// Warning creates a Warning severity diagnostic.
func Warning(summary string, detail string, opts ...Option) IDiagnostic {
	d := &diagnostic{
		severity: SEVERITY_Warning,
		summary:  summary,
		detail:   detail,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(d)
		}
	}
	return d
}

// Info creates an Information severity diagnostic.
func Info(summary string, detail string, opts ...Option) IDiagnostic {
	d := &diagnostic{
		severity: SEVERITY_Info,
		summary:  summary,
		detail:   detail,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(d)
		}
	}
	return d
}

// Verbose creates a Verbose severity diagnostic.
func Verbose(summary string, detail string, opts ...Option) IDiagnostic {
	d := &diagnostic{
		severity: SEVERITY_Verbose,
		summary:  summary,
		detail:   detail,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(d)
		}
	}
	return d
}
