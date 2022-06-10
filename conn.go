package echo

import (
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

var (
	connCnt   int
	connCntMu sync.Mutex
)

type Conn struct {
        Conn       net.Conn       // the connection being wrapped
	files      []io.WriteCloser // optional (open) file to write to
	writers    []*PrettyWriter
	initOnce   sync.Once
	connCnt    int
}

func (c *Conn) AddFileWriter(fname string) *PrettyWriter {
	if c.connCnt == 0 {
		connCntMu.Lock()
		connCnt++
		c.connCnt = connCnt
		connCntMu.Unlock()
	}

	f, err := os.OpenFile(fname, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return &PrettyWriter{err: err}
	}

	w := PrettyWriter{
		Prefix:     fmt.Sprintf("CONN[%d]", c.connCnt),
		writer:     f,
		Verbose:    true,
		LineMax:    120,
		TimeFormat: DefaultTimeFormat,
		Go:         false,
	}

	c.writers = append(c.writers, &w)	
	c.files = append(c.files, f)

	return &w
}

func (c *Conn) AddConsoleWriter() *PrettyWriter {
	if c.connCnt == 0 {
		connCntMu.Lock()
		connCnt++
		c.connCnt = connCnt
		connCntMu.Unlock()
	}

	w := PrettyWriter{
		Prefix:     fmt.Sprintf("CONN[%d]", c.connCnt),
		writer:     os.Stderr,
		Verbose:    false,
		LineMax:    120,
		TimeFormat: DefaultTimeFormat,
		Go:         false,
	}

	c.writers = append(c.writers, &w)	

	return &w
}

func (c *Conn) Write(b []byte) (int, error) {
	c.initOnce.Do(c.initConn)

	for _, w := range c.writers {
		w.Printf("» Write(%d bytes)\n", len(b))
		w.WriteSentBytes(b)
	}

	t := time.Now()

        n, err := c.Conn.Write(b)

	d := time.Since(t)

	for _, w := range c.writers {
		w.Printf("« Write() returned %d,%v — duration: %v\n\n", n, err, d)
	}	

        return n, err
}

func (c *Conn) Read(b []byte) (int, error) {
	c.initOnce.Do(c.initConn)

	for _, w := range c.writers {
		w.Printf("» Read(max %d bytes)\n", len(b))
	}

	t := time.Now()

        n, err := c.Conn.Read(b)

	d := time.Since(t)

	for _, w := range c.writers {
		w.Printf("« Read() returned %d,%v — duration: %v\n", n, err, d)
		w.WriteReceivedBytes(b)
	}	

        return n, err
}

func (c *Conn) Close() error {
	c.initOnce.Do(c.initConn)

	for _, w := range c.writers {
		w.Printf("» Close()\n")
	}

        err := c.Conn.Close()

	for _, w := range c.writers {
		w.Printf("« Close() returned %v\n", err)
	}

	for _, f := range c.files {
		f.Close()
	}

	return err
}

func (c *Conn) LocalAddr() net.Addr {
	c.initOnce.Do(c.initConn)

	for _, w := range c.writers {
		w.Printf("» LocalAddr()\n")
	}

        addr := c.Conn.LocalAddr()

	for _, w := range c.writers {
		w.Printf("« LocalAddr() returned %v\n", addr)
	}

	return addr
}

func (c *Conn) RemoteAddr() net.Addr {
	c.initOnce.Do(c.initConn)

	for _, w := range c.writers {
		w.Printf("» RemoteAddr()\n")
	}

        addr := c.Conn.RemoteAddr()

	for _, w := range c.writers {
		w.Printf("« RemoteAddr() returned %v\n", addr)
	}

	return addr
}

func (c *Conn) SetDeadline(t time.Time) error {
	c.initOnce.Do(c.initConn)

	for _, w := range c.writers {
		w.Printf("» SetDeadline(%s)\n", t.Format(w.TimeFormat))
	}

        err := c.Conn.SetDeadline(t)

	for _, w := range c.writers {
		w.Printf("« SetDeadline() returned %v\n", err)
	}

	return err
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	c.initOnce.Do(c.initConn)

	for _, w := range c.writers {
		w.Printf("» SetReadDeadline(%s)\n", t.Format(w.TimeFormat))
	}

        err := c.Conn.SetReadDeadline(t)

	for _, w := range c.writers {
		w.Printf("« SetReadDeadline() returned %v\n", err)
	}

	return err
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	c.initOnce.Do(c.initConn)

	for _, w := range c.writers {
		w.Printf("» SetWriteDeadline(%s)\n", t.Format(w.TimeFormat))
	}

        err := c.Conn.SetWriteDeadline(t)

	for _, w := range c.writers {
		w.Printf("« SetWriteDeadline() returned %v\n", err)
	}

	return err
}

func (c *Conn) initConn() {
	if len(c.writers) == 0 {
		c.AddConsoleWriter()

		return
	}

	for _, w := range c.writers {
		if w.TimeFormat == "" {
			w.TimeFormat = DefaultTimeFormat
		}
	}
}
