package diagnostic

import (
	"strconv"
	"strings"
)

type Frame struct {
	PC       uintptr
	Func     string
	File     string
	Line     int
	OffsetPC uintptr
}

func (t Frame) String() string {
	sb := strings.Builder{}

	// Formating as so:
	//     <function <str>> <program counter <hex>>
	//         <file <str>>:<line <dec>> <program counter offset <hex>>
	sb.WriteString(t.Func)
	sb.WriteString(" 0x")
	sb.WriteString(strconv.FormatUint(uint64(t.PC), 16))
	sb.WriteByte('\n')

	sb.WriteByte('\t')
	sb.WriteString(t.File)
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(t.Line))
	sb.WriteString(" +0x")
	sb.WriteString(strconv.FormatUint(uint64(t.OffsetPC), 16))
	sb.WriteByte('\n')

	return sb.String()
}
