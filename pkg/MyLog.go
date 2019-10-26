// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package log implements a simple logging package. It defines a type, Logger,
// with methods for formatting output. It also has a predefined 'standard'
// Logger accessible through helper functions Print[f|ln], Fatal[f|ln], and
// Panic[f|ln], which are easier to use than creating a Logger manually.
// That logger writes to standard error and prints the date and time
// of each logged message.
// Every log message is output on a separate line: if the message being
// printed does not end in a newline, the logger will add one.
// The Fatal functions call os.Exit(1) after writing the log message.
// The Panic functions call panic after writing the log message.
package pkg

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"time"
)

const (
	Ldate         = 1 << iota // the date in the local time zone: 2009/01/23
	Ltime                     // the time in the local time zone: 01:23:23
	Lmicroseconds             // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile                 // full file name and line number: /a/b/c/d.go:23
	Lshortfile                // final file name element and line number: d.go:23. overrides Llongfile
	LUTC                      // if Ldate or Ltime is set, use UTC rather than the local time zone
	Ldebug
	Linfo
	Lwarn
	Lerror
	Lpanic
	Ltrace
	LstdFlags = Ldate | Ltime | Llongfile // initial values for the standard logger

)

type Logger struct {
	mu     sync.Mutex // ensures atomic writes; protects the following fields
	prefix string     // prefix to write at beginning of each line
	flag   int        // properties
	out    io.Writer  // destination for output
	buf    []byte
}


func New(out io.Writer, prefix string, flag int) *Logger {
	return &Logger{out: out, prefix: prefix, flag: flag}
}

var std = New(os.Stderr, "", LstdFlags)

// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
}

// formatHeader writes log header to buf in following order:
//   * l.prefix (if it's not blank),
//   * date and/or time (if corresponding flags are provided),
//   * file and line number (if corresponding flags are provided).
func (l *Logger) formatHeader(buf *[]byte, t time.Time, file string, line int, level int) {
	*buf = append(*buf, l.prefix...)
	if l.flag&(Ldate|Ltime|Lmicroseconds) != 0 {
		if l.flag&LUTC != 0 {
			t = t.UTC()
		}
		if l.flag&Ldate != 0 {
			year, month, day := t.Date()
			itoa(buf, year, 4)
			*buf = append(*buf, '/')
			itoa(buf, int(month), 2)
			*buf = append(*buf, '/')
			itoa(buf, day, 2)
			*buf = append(*buf, ' ')
		}
		if l.flag&(Ltime|Lmicroseconds) != 0 {
			hour, min, sec := t.Clock()
			itoa(buf, hour, 2)
			*buf = append(*buf, ':')
			itoa(buf, min, 2)
			*buf = append(*buf, ':')
			itoa(buf, sec, 2)
			if l.flag&Lmicroseconds != 0 {
				*buf = append(*buf, '.')
				itoa(buf, t.Nanosecond()/1e3, 6)
			}
			*buf = append(*buf, ' ')
		}
	}

	if l.flag&(Ldebug|Linfo|Lwarn|Lerror|Lpanic) != 0 {
		switch level {
		case Ldebug:
			*buf = append(*buf, []byte(" DEBUG ")...)
		case Linfo:
			*buf = append(*buf, []byte(" INFO ")...)
		case Lwarn:
			*buf = append(*buf, []byte(" WARN ")...)
		case Lerror:
			*buf = append(*buf, []byte(" ERROR ")...)
		case Lpanic:
			*buf = append(*buf, []byte(" PANIC ")...)
		case Ltrace:
			*buf = append(*buf, []byte(" TRACE ")...)
		default:
			*buf = append(*buf, []byte(" UNKNOWN ")...)
		}
	}

	if l.flag&(Lshortfile|Llongfile) != 0 {
		if l.flag&Lshortfile != 0 {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
		}
		*buf = append(*buf, file...)
		*buf = append(*buf, ':')
		itoa(buf, line, -1)
		*buf = append(*buf, ": "...)
	}
}

func (l *Logger) Output(calldepth int, level int, s string) error {
	now := time.Now() // get this early.
	var file string
	var line int
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.flag&(Ldebug|Linfo|Lwarn|Lerror|Lpanic|Ltrace) != 0 {
		if l.flag&Linfo != 0 && level < Linfo {
			return nil
		} else if l.flag&Lwarn != 0 && level < Lwarn {
			return nil
		} else if l.flag&Lerror != 0 && level < Lerror {
			return nil
		} else if l.flag&Lpanic != 0 && level < Lpanic {
			return nil
		} else if l.flag&Ltrace != 0 && level < Ltrace {
			return nil
		}
	}
	if l.flag&(Lshortfile|Llongfile) != 0 {
		// Release lock while getting caller info - it's expensive.
		l.mu.Unlock()
		var ok bool
		_, file, line, ok = runtime.Caller(calldepth)
		if !ok {
			file = "???"
			line = 0
		}
		l.mu.Lock()
	}
	l.buf = l.buf[:0]
	l.formatHeader(&l.buf, now, file, line, level)
	l.buf = append(l.buf, s...)
	if len(s) == 0 || s[len(s)-1] != '\n' {
		l.buf = append(l.buf, '\n')
	}
	_, err := l.out.Write(l.buf)
	return err
}

// Printf calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Output(2, Linfo, fmt.Sprintf(format, v...))
}

// Print calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Print.
func (l *Logger) Print(v ...interface{}) { l.Output(2, Linfo, fmt.Sprint(v...)) }

// Println calls l.Output to print to the logger.
// Arguments are handled in the manner of fmt.Println.
func (l *Logger) Println(v ...interface{}) { l.Output(2, Linfo, fmt.Sprintln(v...)) }

// Fatal is equivalent to l.Print() followed by a call to os.Exit(1).
func (l *Logger) Fatal(v ...interface{}) {
	l.Output(2, Lerror, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to l.Printf() followed by a call to os.Exit(1).
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Output(2, Lerror, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln is equivalent to l.Println() followed by a call to os.Exit(1).
func (l *Logger) Fatalln(v ...interface{}) {
	l.Output(2, Lerror, fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic is equivalent to l.Print() followed by a call to panic().
func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.Output(2, Lpanic, s)
	panic(s)
}

// Panicf is equivalent to l.Printf() followed by a call to panic().
func (l *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.Output(2, Lpanic, s)
	panic(s)
}

// Panicln is equivalent to l.Println() followed by a call to panic().
func (l *Logger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.Output(2, Lpanic, s)
	panic(s)
}

// Flags returns the output flags for the logger.
func (l *Logger) Flags() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.flag
}

// SetFlags sets the output flags for the logger.
func (l *Logger) SetFlags(flag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flag = flag
}

// Prefix returns the output prefix for the logger.
func (l *Logger) Prefix() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.prefix
}

// SetPrefix sets the output prefix for the logger.
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// SetOutput sets the output destination for the standard logger.
func SetOutput(w io.Writer) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.out = w
}

// Flags returns the output flags for the standard logger.
func Flags() int {
	return std.Flags()
}

// SetFlags sets the output flags for the standard logger.
func SetFlags(flag int) {
	std.SetFlags(flag)
}

// Prefix returns the output prefix for the standard logger.
func Prefix() string {
	return std.Prefix()
}

// SetPrefix sets the output prefix for the standard logger.
func SetPrefix(prefix string) {
	std.SetPrefix(prefix)
}

// These functions write to the standard logger.

// Print calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Print.
func Print(v ...interface{}) {
	std.Output(2, Linfo, fmt.Sprint(v...))
}

// Printf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Printf(format string, v ...interface{}) {
	std.Output(2, Linfo, fmt.Sprintf(format, v...))
}

// Println calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Println.
func Println(v ...interface{}) {
	std.Output(2, Linfo, fmt.Sprintln(v...))
}

// Fatal is equivalent to Print() followed by a call to os.Exit(1).
func Fatal(v ...interface{}) {
	std.Output(2, Lerror, fmt.Sprint(v...))
	os.Exit(1)
}

// Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
func Fatalf(format string, v ...interface{}) {
	std.Output(2, Lerror, fmt.Sprintf(format, v...))
	os.Exit(1)
}

// Fatalln is equivalent to Println() followed by a call to os.Exit(1).
func Fatalln(v ...interface{}) {
	std.Output(2, Lerror, fmt.Sprintln(v...))
	os.Exit(1)
}

// Debug calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Debug(v ...interface{}) {
	l.Output(2, Ldebug, fmt.Sprint(v...))
}

// Debugf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.Output(2, Ldebug, fmt.Sprintf(format, v...))
}

// Debugln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Debugln(v ...interface{}) {
	l.Output(2, Ldebug, fmt.Sprintln(v...))
}

// Info calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Info(v ...interface{}) {
	l.Output(2, Linfo, fmt.Sprint(v...))
}

// Infof calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Infof(format string, v ...interface{}) {
	l.Output(2, Linfo, fmt.Sprintf(format, v...))
}

// Infoln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Infoln(v ...interface{}) {
	l.Output(2, Linfo, fmt.Sprintln(v...))
}

// Warn calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Warn(v ...interface{}) {
	l.Output(2, Lwarn, fmt.Sprint(v...))
}

// Warnf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Output(2, Lwarn, fmt.Sprintf(format, v...))
}

// Warnln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Warnln(v ...interface{}) {
	l.Output(2, Lwarn, fmt.Sprintln(v...))
}

// Error calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Error(v ...interface{}) {
	l.Output(2, Lerror, fmt.Sprint(v...))
}

// Errorf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Output(2, Lerror, fmt.Sprintf(format, v...))
}

// Errorln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Errorln(v ...interface{}) {
	l.Output(2, Lerror, fmt.Sprintln(v...))
}

// Trace calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Trace(v ...interface{}) {
	l.Output(2, Ltrace, fmt.Sprint(v...))
}

// Tracef calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Tracef(format string, v ...interface{}) {
	l.Output(2, Ltrace, fmt.Sprintf(format, v...))
}

// Traceln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func (l *Logger) Traceln(v ...interface{}) {
	l.Output(2, Ltrace, fmt.Sprintln(v...))
}

// Debug calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Debug(v ...interface{}) {
	std.Output(2, Ldebug, fmt.Sprint(v...))
}

// Debugf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Debugf(format string, v ...interface{}) {
	std.Output(2, Ldebug, fmt.Sprintf(format, v...))
}

// Debugln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Debugln(v ...interface{}) {
	std.Output(2, Ldebug, fmt.Sprintln(v...))
}

// Info calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Info(v ...interface{}) {
	std.Output(2, Linfo, fmt.Sprint(v...))
}

// Infof calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Infof(format string, v ...interface{}) {
	std.Output(2, Linfo, fmt.Sprintf(format, v...))
}

// Infoln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Infoln(v ...interface{}) {
	std.Output(2, Linfo, fmt.Sprintln(v...))
}

// Warn calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Warn(v ...interface{}) {
	std.Output(2, Lwarn, fmt.Sprint(v...))
}

// Warnf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Warnf(format string, v ...interface{}) {
	std.Output(2, Lwarn, fmt.Sprintf(format, v...))
}

// Warnln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Warnln(v ...interface{}) {
	std.Output(2, Lwarn, fmt.Sprintln(v...))
}

// Error calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Error(v ...interface{}) {
	std.Output(2, Lerror, fmt.Sprint(v...))
}

// Errorf calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Errorf(format string, v ...interface{}) {
	std.Output(2, Lerror, fmt.Sprintf(format, v...))
}

// Errorln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Errorln(v ...interface{}) {
	std.Output(2, Lerror, fmt.Sprintln(v...))
}

// Trace calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Trace(v ...interface{}) {
	std.Output(2, Ltrace, fmt.Sprint(v...))
}

// Tracef calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Tracef(format string, v ...interface{}) {
	std.Output(2, Ltrace, fmt.Sprintf(format, v...))
}

// Traceln calls Output to print to the standard logger.
// Arguments are handled in the manner of fmt.Printf.
func Traceln(v ...interface{}) {
	std.Output(2, Ltrace, fmt.Sprintln(v...))
}

// Panic is equivalent to Print() followed by a call to panic().
func Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	std.Output(2, Lpanic, s)
	panic(s)
}

// Panicf is equivalent to Printf() followed by a call to panic().
func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	std.Output(2, Lpanic, s)
	panic(s)
}

// Panicln is equivalent to Println() followed by a call to panic().
func Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	std.Output(2, Lpanic, s)
	panic(s)
}

// Output writes the output for a logging event. The string s contains
// the text to print after the prefix specified by the flags of the
// Logger. A newline is appended if the last character of s is not
// already a newline. Calldepth is the count of the number of
// frames to skip when computing the file name and line number
// if Llongfile or Lshortfile is set; a value of 1 will print the details
// for the caller of Output.
func Output(calldepth int, level int, s string) error {
	return std.Output(calldepth+1, level, s) // +1 for this frame.
}