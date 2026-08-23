package diagnostic

type opt int
type option interface {
	TypeName() string
}

const (
	opttype_Category     = "Category"
	opttype_DoStacktrace = "DoStacktrace"
)

const (
	OPTION_DOSTACKTRACE_FALSE doStacktrace = 0
	OPTION_DOSTACKTRACE_TRUE  doStacktrace = 1
)

type doStacktrace opt

func (t doStacktrace) TypeName() string {
	return opttype_DoStacktrace
}

type category string

const (
	Category_InvalidParamter category = "Invalid Parameter"
)
