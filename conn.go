package telnet

import "io"

type connection struct {
	in     io.Reader
	out    io.Writer
	opts   *optionMap
	rawOut io.Writer
}

func newConnection(r io.Reader, w io.Writer) *connection {
	return &connection{
		in:     NewReader(r),
		opts:   newOptionMap(),
		out:    NewWriter(w),
		rawOut: w,
	}
}

func (c *connection) Read(p []byte) (n int, err error) {
	n, err = c.in.Read(p)
	switch t := err.(type) {
	case *telnetGoAhead:
		err = nil
	case *telnetOptionCommand:
		err = nil
		opt := c.opts.get(byte(t.opt))
		opt.receive(byte(t.cmd), func(cmd byte) {
			_, err = c.rawOut.Write([]byte{IAC, cmd, byte(t.opt)})
		})
	}
	return
}

func (c *connection) Write(p []byte) (n int, err error) {
	n, err = c.out.Write(p)
	if err == nil {
		_, err = c.rawOut.Write([]byte{IAC, GA})
	}
	return
}
