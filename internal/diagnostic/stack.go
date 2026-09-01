package diagnostic

import (
	"io"
	"runtime"
)

type Stack []Frame

func (t Stack) Write(w io.Writer) {
	for i := len(t) - 1; i >= 0; i-- {
		t[i].Write(w)
	}
}

func (t Stack) String() string {
	buf := make([]byte, 0, len(t)*112)

	for i := len(t) - 1; i >= 0; i-- {
		buf = t[i].AppendTo(buf)
	}

	return string(buf)
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
