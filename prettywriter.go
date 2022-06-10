package echo

import (
	"fmt"
	"io"
	"time"
)

const DefaultTimeFormat = "2006-01-02 15:04:05.00000-0800"

type PrettyWriter struct {
	TimeFormat string    // format of the timestamp prefix
	writer     io.Writer // the writer to write to
	Go         bool      // whether to print []byte slices as Go code
	Prefix     string    // string to prefix each line with
	LineMax    int       // maximum length of lines (must be at least 80, because vt100); default 120
	Verbose    bool      // print the data read/written
	sendCnt    int
	recvCnt    int
	err        error
}

func (p *PrettyWriter) Err() error {
	return p.err
}

func (p *PrettyWriter) ts() string {
	return time.Now().Format(p.TimeFormat)
}

func (p *PrettyWriter) Printf(format string, args ...interface{}) {
	if p.Go {
		_, err := fmt.Fprintf(p.writer, "// ")
		if err != nil {
			p.err = err
			return
		}
	}

	_, err := fmt.Fprintf(p.writer, "%s %s ", p.Prefix, p.ts())
	if err != nil {
		p.err = err
		return
	}
	
	_, err = fmt.Fprintf(p.writer, format, args...)
	if err != nil {
		p.err = err
		return
	}
}

func hex(b byte) byte {
	b = b%16

	if b < 10 {
		return b + '0'
	}
	
	return b + 'A' - 10
}

func printable(b byte) byte {
	const trns ="................" +
		"................" +
		" !\"#$%&'()*+,-./" +
		"0123456789:;<=>?" +
		"@ABCDEFGHIJKLMNO" +
		"PQRSTUVWXYZ[\\]^_" +
		"`abcdefghijklmno" +
		"pqrstuvwxyz{|}~." +
		"................" +
		"................" +
		"................" +
		"................" +
		"................" +
		"................" +
		"................" +
		"................"
	return trns[int(b)]
}

func (p *PrettyWriter) WriteGoDump(varName string, b []byte) {
	_, err := fmt.Fprintf(p.writer, "var %s = []byte{\n", varName)
	if err != nil {
		p.err = err
		return
	}

	start := 0
	for start < len(b) {
		end := start + 16
		if end > len(b) {
			end = len(b)
		}

		line := make([]byte, 0, 120)

		for _, c := range b[start:end] {
			line = append(line, ' ', '0', 'x', hex(c/16), hex(c%16), ',')
		}

		_, err = fmt.Fprintf(p.writer, "       %s\n", line)
		if err != nil {
			p.err = err
			return
		}

		start = end
	}
	_, err = fmt.Fprintf(p.writer, "}\n\n")
	if err != nil {
		p.err = err
		return
	}
}

func (p *PrettyWriter) WriteSentBytes(b []byte) {
	p.sendCnt++

	if p.Go {
		p.WriteGoDump(fmt.Sprintf("wbuf%00d", p.sendCnt), b)		
	} else {
		p.WriteBytesHexDump(b)
	}
}

func (p *PrettyWriter) WriteReceivedBytes(b []byte) {
	p.recvCnt++

	if p.Go {
		p.WriteGoDump(fmt.Sprintf("rbuf%00d", p.recvCnt), b)		
	} else {
		p.WriteBytesHexDump(b)
	}
}

func (p *PrettyWriter) WriteBytesHexDump(b []byte) {
	if !p.Verbose {
		return
	}

	lineMax := p.LineMax
	if lineMax == 0 {
		lineMax = 115
	}
	if lineMax < 80 {
		lineMax = 80
	}

	bytesPerLine := (lineMax - 4 - len(p.Prefix)) / 4
	for bytesPerLine / 10 + bytesPerLine * 4 + 4 + len(p.Prefix) > lineMax {
		bytesPerLine--
	}

	if bytesPerLine < 5 {
		bytesPerLine = 5
	}
	start := 0

	for start < len(b) {
		end := start + bytesPerLine
		if end > len(b) {
			end = len(b)
		}

		line := make([]byte, 0, lineMax)
		line = append(line, ' ')

		for i, c := range b[start:end] {
			line = append(line, hex(c/16), hex(c%16), ' ')
			if i % 10 == 9 && i != end-start-1 {
				line = append(line, ' ')
			}
		}

		for i := bytesPerLine + start - end; i > 0; i-- {
			line = append(line, ' ', ' ', ' ')
			if i % 10 == 9 && i != 1 {
				line = append(line, ' ')
			}
		}

		line = append(line, ' ', '|', ' ')

		for _, c := range b[start:end] {
			line = append(line, printable(c))
		}

		_, err := fmt.Fprintf(p.writer, "%s%s\n", p.Prefix, string(line))
		if err != nil {
			p.err = err
			return
		}

		start = end
	}
}
