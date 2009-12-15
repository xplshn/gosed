//
//  h_cmd.go
//  sed
//
// Copyright (c) 2009 Geoffrey Clements
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//

package sed

import (
	"bytes"
	"fmt"
	"os"
)

type h_cmd struct {
	addr	*address
	replace	bool
}

func (c *h_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *h_cmd) String() string {
	if c != nil {
		if c.addr != nil {
			if c.replace {
				return fmt.Sprint("{Replace hold space with contents of pattern space Cmd addr:%s}", c.addr.String())
			} else {
				return fmt.Sprint("{Append a newline and the pattern space to the hold space Cmd addr:%s}", c.addr.String())
			}
		} else {
			if c.replace {
				return fmt.Sprint("{Replace hold space with contents of pattern space Cmd}")
			} else {
				return fmt.Sprint("{Append a newline and the pattern space to the hold space Cmd")
			}
		}
	}

	return fmt.Sprint("{Append/Replace hold space with contents of pattern space}")
}

func (c *h_cmd) processLine(s *Sed) (bool, os.Error) {
	if c.replace {
		s.holdSpace = copyByteSlice(s.patternSpace)
	} else {
		s.holdSpace = bytes.AddByte(s.holdSpace, '\n')
		s.holdSpace = bytes.Add(s.holdSpace, s.patternSpace)
	}
	return false, nil
}

func NewHCmd(pieces [][]byte, addr *address) (*h_cmd, os.Error) {
	if len(pieces) > 1 {
		return nil, WrongNumberOfCommandParameters
	}
	cmd := new(h_cmd)
	if pieces[0][0] == 'h' {
		cmd.replace = true
	}
	cmd.addr = addr
	return cmd, nil
}
