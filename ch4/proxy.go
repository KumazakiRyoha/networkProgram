package main

import "io"

func proxy(from io.Reader, to io.Writer) error {
	fromWriter, fromIsWriter := from.(io.Writer)
	toReader, toIsReader := to.(io.Reader)
	if toIsReader && fromIsWriter {
		// Send replies since "from" and "to" implement the necessary interface
		go func() {
			_, _ = io.Copy(fromWriter, toReader)
		}()
	}
	_, err := io.Copy(to, from)
	return err
}
