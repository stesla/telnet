package telnet

import "io"

type connection struct {
	io.Reader
	io.Writer

	out io.Writer
}

func newConnection(r io.Reader, w io.Writer) *connection {
	return &connection{
		Reader: NewReader(r),
		Writer: NewWriter(w),
		out:    w,
	}
}

func (c *connection) Read(p []byte) (n int, err error) {
	n, err = c.Reader.Read(p)
	switch t := err.(type) {
	case *telnetGoAhead:
		err = nil
	case *telnetOptionCommand:
		var cmd byte
		switch t.cmd {
		case DO, DONT:
			cmd = WONT
		case WILL, WONT:
			cmd = DONT
		}
		_, err = c.out.Write([]byte{IAC, cmd, byte(t.opt)})
	}
	return
}
