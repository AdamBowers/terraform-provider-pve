package diagnostic

type Option func(*diagnostic)

// WithCategory sets a specific category on the diagnostic.
func WithCategory(cat category) Option {
	return func(d *diagnostic) {
		d.category = cat
	}
}

// WithStacktrace will generate the stacktrace on the diagnostic. Do not use in conjunction with: WithStacktraceAbsDepth,
// WithStacktraceDepth, WithoutStacktrace
func WithStacktrace() Option {
	return func(d *diagnostic) {
		d.stacktrace = GenerateStack(4) // 4 is to skip runtime.callers, Generate Stack, this function and the diagnostic contructor
	}
}

// WithStacktraceAbsDepth will generate the stacktrace on a the diagnostic with an ability to specify an custom
// abolute skip amount of the stack. i.e. allows the capturing of helper methods.
// Do noy use in conjunction with: WithStacktrace, WithStacktraceDepth, WithoutStacktrace
func WithStacktraceAbsDepth(skip int) Option {
	return func(d *diagnostic) {
		d.stacktrace = GenerateStack(skip)
	}
}

// WithStacktraceDepth will generate the stacktrace on a the diagnostic with an ability to specify an custom
// abolute skip amount of the stack. i.e. allows the capturing of helper methods.
// Do noy use in conjunction with: WithStacktrace, WithStacktraceAbsDepth, WithoutStacktrace
func WithStacktraceDepth(skip int) Option {
	return func(d *diagnostic) {
		d.stacktrace = GenerateStack(skip + 4) // 4 is to skip runtime.callers, Generate Stack, this function and the diagnostic contructor
	}
}

// WithoutStacktrace initialises an empty trace.
// Use this with Error diagnostic to disable the automatic generation of stacktrace
// Do noy use in conjunction with: WithStacktrace, WithStacktraceAbsDepth, WithStacktraceDepth
func WithoutStacktrace() Option {
	return func(d *diagnostic) {
		d.stacktrace = Stack{}
	}
}
