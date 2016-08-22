
package utils

import (
	"io"
)

// Writer for one line, compatible with csv.Writer interface on Write
type CgrRecordWriter interface {
	Write([]string) error
	Flush()
}

func NewCgrIORecordWriter(w io.Writer) *CgrIORecordWriter {
	return &CgrIORecordWriter{w: w}
}

type CgrIORecordWriter struct {
	w io.Writer
}

func (self *CgrIORecordWriter) Write(record []string) error {
	for _, fld := range append(record, "\n") { // Postpend the new line char and write record in the writer
		if _, err := io.WriteString(self.w, fld); err != nil {
			return err
		}
	}
	return nil
}

// ToDo: make sure we properly handle this method
func (self *CgrIORecordWriter) Flush() {
	return
}
