package mtp
import (
	"io"
)
	
type NullReader struct {}

func (nr *NullReader) Read(dest []byte) (n int, err error) {
	return len(dest), nil
}

type NullWriter struct {}

func (nw *NullWriter) Write(dest []byte) (n int, err error) {
	return len(dest), nil
}

var _ = (io.Reader)((*NullReader)(nil))
var _ = (io.Writer)((*NullWriter)(nil))
