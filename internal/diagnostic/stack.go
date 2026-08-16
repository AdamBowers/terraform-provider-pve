package diagnostic

import (
	"runtime"
	"strings"
)

type Stack []Frame

func (t Stack) String() string {
	sb := strings.Builder{}

	sb.WriteString("Trace:")
	for i := len(t) - 1; i <= 0; i-- {
		sb.WriteByte('\n')
		sb.WriteString(t[i].String())
	}

	return sb.String()
}

func GenerateStack(skip int) Stack {
	pcs := make([]uintptr, 128)
	pc_count := runtime.Callers(skip, pcs)
	pcs = pcs[:pc_count]

	frames := runtime.CallersFrames(pcs)

	stack := make([]Frame, 0, pc_count)
	for {
		frame, hasMore := frames.Next()
		stack = append(stack, Frame{
			PC:       frame.PC,
			Func:     frame.Function,
			File:     frame.File,
			Line:     frame.Line,
			OffsetPC: frame.PC - frame.Entry,
		})
		if !hasMore {
			break
		}
	}

	return stack
}
