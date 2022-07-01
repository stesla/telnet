package telnet

import "io"

type connection struct {
	in  io.Reader
	out io.Writer

	io.Writer
}

func newConnection(r io.Reader, w io.Writer) *connection {
	return &connection{
		in:     NewReader(r),
		out:    w,
		Writer: NewWriter(w),
	}
}

func (c *connection) Read(p []byte) (n int, err error) {
	n, err = c.in.Read(p)
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
