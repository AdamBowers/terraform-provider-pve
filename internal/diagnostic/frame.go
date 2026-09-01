package diagnostic

import (
	"io"
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

func (t Frame) Write(w io.Writer) {
	var buf [32]byte

	// Format:
	//     <function <str>> <program counter <hex>>
	//         <file <str>>:<line <dec>> <program counter offset <hex>>

	io.WriteString(w, t.Func)
	io.WriteString(w, " 0x")

	pc := strconv.AppendUint(buf[:0], uint64(t.PC), 16)
	w.Write(pc)

	io.WriteString(w, "\n\t")
	io.WriteString(w, t.File)
	io.WriteString(w, ":")

	ln := strconv.AppendInt(buf[:0], int64(t.Line), 10)
	w.Write(ln)

	io.WriteString(w, " +0x")

	opc := strconv.AppendUint(buf[:0], uint64(t.OffsetPC), 16)
	w.Write(opc)

	io.WriteString(w, "\n")
}

func (t Frame) AppendTo(buf []byte) []byte {
	buf = append(buf, t.Func...)
	buf = append(buf, " 0x"...)
	buf = strconv.AppendUint(buf, uint64(t.PC), 16)

	buf = append(buf, '\n', '\t')
	buf = append(buf, t.File...)
	buf = append(buf, ':')
	buf = strconv.AppendInt(buf, int64(t.Line), 10)
	buf = append(buf, " +0x"...)
	buf = strconv.AppendUint(buf, uint64(t.OffsetPC), 16)
	buf = append(buf, '\n')

	return buf
}

func (t Frame) String() string {
	var b strings.Builder
	b.Grow(256)
	t.Write(&b)
	return b.String()
}
