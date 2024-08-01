// cmd.go
// sed
//
// Original code: Copyright (c) 2009 Geoffrey Clements (MIT License)
// Modified code: Copyright (c) 2024 xplshn (3BSD License)
// For details, see the [LICENSE](https://github.com/xplshn/gosed) file at the root directory of this project
package sed

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// Err definitions
var (
	ErrWrongNumberOfCommandParameters = errors.New("Wrong number of parameters for command")
	ErrUnknownScriptCommand           = errors.New("Unknown script command")
	ErrInvalidSCommandFlag            = errors.New("Invalid flag for s command")
	ErrRegularExpressionExpected      = errors.New("Expected a regular expression, got zero length string")
	ErrUnterminatedRegularExpression  = errors.New("Unterminated regular expression")
	ErrNoSupportForTwoAddress         = errors.New("This command doesn't support an address range or to end of file")
	ErrNotImplemented                 = errors.New("This command command hasn't been implemented yet")
)

// Cmd represents a command that can be executed by the Sed processor // It includes methods for processing lines and converting the command to a string.
type Cmd interface {
	fmt.Stringer
	processLine(s *Sed) (stop bool, err error)
}

// Address represents a pattern or address that can be matched against a line of input // It includes a method to determine if the address matches the given line and line number.
type Address interface {
	match(line []byte, lineNumber int) bool
}

const (
	addressLine = iota
	addressRange
	addressToEndOfFile
	addressLastLine
	addressRegEx
)

type address struct {
	not         bool
	addressType int
	rangeStart  int
	rangeEnd    int
	regex       *regexp.Regexp
}

func (a *address) getTypeAsString() string {
	if a != nil {
		switch a.addressType {
		case addressLine:
			return "addressLine"
		case addressRange:
			return "addressRange"
		case addressToEndOfFile:
			return "addressToEndOfFile"
		case addressLastLine:
			return "addressLastLine"
		case addressRegEx:
			return "addressRegEx"
		default:
			return "ADDRESS_UNKNOWN"
		}
	}
	return "nil"
}

func (a *address) String() string {
	return fmt.Sprintf("address{type: %s rangeStart:%d rangeEnd:%d regex:%v}", a.getTypeAsString(), a.rangeStart, a.rangeEnd, a.regex)
}

func (a *address) match(line []byte, lineNumber int) bool {
	val := true
	if a != nil {
		switch a.addressType {
		case addressLine:
			val = lineNumber == a.rangeStart
		case addressRange:
			val = lineNumber >= a.rangeStart && lineNumber <= a.rangeEnd
		case addressToEndOfFile:
			val = lineNumber >= a.rangeStart
		case addressLastLine:
			val = false // this is wrong!
		case addressRegEx:
			val = a.regex.Match(line)
		default:
			val = false
		}
		if a.not {
			val = !val
		}
	}
	return val
}

func getNumberFromLine(s []byte) ([]byte, int, error) {
	idx := 0
	for {
		if s[idx] < '0' || s[idx] > '9' {
			break
		}
		idx++
	}
	i, err := strconv.Atoi(string(s[0:idx]))
	if err != nil {
		return s, -1, err
	}
	return s[idx:], i, nil
}

// A nil address means match any line
func checkForAddress(s []byte) ([]byte, *address, error) {
	var err error
	if s[0] == '/' {
		// regular expression address
		s = s[1:]
		idx := bytes.IndexByte(s, '/')
		if idx < 0 {
			return s, nil, ErrUnterminatedRegularExpression
		}
		r := s[0:idx]
		if len(r) == 0 {
			return s, nil, ErrRegularExpressionExpected
		}
		// s is now just the command
		s = s[idx+1:]
		addr := new(address)
		addr.addressType = addressRegEx
		addr.regex, err = regexp.CompilePOSIX(string(r))
		if err != nil {
			return s, nil, err
		}
		return s, addr, nil
	} else if s[0] == '$' {
		// end of file
		addr := new(address)
		addr.addressType = addressLastLine
		// s is now just the command
		s = s[1:]
		return s, addr, nil
	} else if s[0] >= '0' && s[0] <= '9' {
		// numeric line address
		addr := new(address)
		addr.addressType = addressLine
		s, addr.rangeStart, err = getNumberFromLine(s)
		if err != nil {
			return s, nil, err
		}
		addr.rangeEnd = addr.rangeStart
		if s[0] == ',' {
			s = s[1:]
			if len(s) > 0 && s[0] >= '0' && s[0] <= '9' {
				addr.addressType = addressRange
				s, addr.rangeEnd, err = getNumberFromLine(s)
				if err != nil {
					return s, nil, err
				}
				// if end range is less than start only match single line
				if addr.rangeEnd < addr.rangeStart {
					addr.addressType = addressLine
					addr.rangeEnd = 0
				}
			} else {
				addr.addressType = addressToEndOfFile
			}
		}
		if s[0] == '!' {
			addr.not = true
			s = s[1:]
		}
		return s, addr, nil
	}
	return s, nil, nil
}

// NewCmd creates a new Cmd instance based on the given Sed object and line of input.
// It parses the line to determine the appropriate command type and returns an instance
// of the corresponding command. It also processes any addresses specified in the line.
//
// Parameters:
//
//	s - The Sed object used to execute commands.
//	line - The line of input that specifies the command and address.
//
// Returns:
//
//	A Cmd instance corresponding to the specified command or an error if the command is unknown.
func NewCmd(s *Sed, line []byte) (Cmd, error) {

	var err error
	var addr *address
	line, addr, err = checkForAddress(line)
	if err != nil {
		return nil, err
	}

	if len(line) > 0 {
		switch line[0] {
		case 'a':
			return NewACmd(s, line, addr)
		case 'b':
			return NewBCmd(bytes.Split(line, []byte{'/'}), addr)
		case 'c':
			return NewCCmd(s, line, addr)
		case 'd', 'D':
			return NewDCmd(bytes.Split(line, []byte{'/'}), addr)
		case 'g', 'G':
			return NewGCmd(bytes.Split(line, []byte{'/'}), addr)
		case 'h', 'H':
			return NewHCmd(bytes.Split(line, []byte{'/'}), addr)
		case 'i':
			return NewICmd(s, line, addr)
		case 'n':
			return NewNCmd(bytes.Split(line, []byte{'/'}), addr)
		case 'P', 'p':
			return NewPCmd(bytes.Split(line, []byte{'/'}), addr)
		case 'q':
			return NewQCmd(bytes.Split(line, []byte{'/'}), addr)
		case 'r':
			return NewRCmd(line, addr)
		case 's':
			return NewSCmd(bytes.Split(line, []byte{'/'}), addr)
		case '=':
			return NewEqlCmd(bytes.Split(line, []byte{'/'}), addr)
		case 'x':
			return NewXCmd(bytes.Split(line, []byte{'/'}), addr)
		}
	}
	return nil, ErrUnknownScriptCommand
}
