// A_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(1)a%5C,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20next%20input%20line. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(1)a%5C,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20next%20input%20line.
// What you gotta do with us (.gopart files) is: `cat ./definitions/*.gopart > ./commands.go`, that way you can compile `sed`.
package sed

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"io"
)

type a_cmd struct {
	addr *address
	text []byte
}

func (c *a_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *a_cmd) String() string {
	if c != nil {
		if c.addr != nil {
			return fmt.Sprintf("{a command addr:%s text:%s}", c.addr.String(), c.text)
		}
		return fmt.Sprintf("{a command text:%s}", c.text)
	}
	return fmt.Sprintf("{a command}")
}

func (c *a_cmd) processLine(s *Sed) (bool, error) {
	return false, nil
}

func NewACmd(s *Sed, line []byte, addr *address) (*a_cmd, error) {
	cmd := new(a_cmd)
	cmd.addr = addr
	cmd.text = line[1:]
	for bytes.HasSuffix(cmd.text, []byte{'\\'}) {
		cmd.text = cmd.text[0 : len(cmd.text)-1]
		line, err := s.getNextScriptLine()
		if err != nil {
			break
		}
		buf := bytes.NewBuffer(cmd.text)
		buf.WriteRune('\n')
		buf.Write(line)
		cmd.text = buf.Bytes()
	}
	cmd.text = trimSpaceFromBeginning(cmd.text)
	return cmd, nil
}
// E-OF: A_CMD //
// B_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)b%20label,the%20end%20of%20the%20script. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)b%20label,the%20end%20of%20the%20script.
type b_cmd struct {
	addr  *address
	label string
}

func (c *b_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *b_cmd) String() string {
	if c != nil {
		if c.addr != nil {
			return fmt.Sprintf("{b command label: %s Cmd addr:%s}", c.label, c.addr.String())
		}
		return fmt.Sprintf("{b command label: %s Cmd}", c.label)
	}
	return fmt.Sprintf("{b command}")
}

func (c *b_cmd) processLine(s *Sed) (bool, error) {
	return true, NotImplemented
}

func NewBCmd(pieces [][]byte, addr *address) (*b_cmd, error) {
	if len(pieces) != 1 {
		return nil, WrongNumberOfCommandParameters
	}
	cmd := new(b_cmd)
	cmd.addr = addr
	cmd.label = string(bytes.TrimSpace(pieces[0][1:]))
	return cmd, nil
}
// E-OF: B_CMD //
// C_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)c%5C,output.%20%20Start%20the%20next%20cycle. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)c%5C,output.%20%20Start%20the%20next%20cycle.
type c_cmd struct {
	addr *address
	text []byte
}

func (c *c_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *c_cmd) String() string {
	if c != nil {
		if c.addr != nil {
			return fmt.Sprintf("{c command addr:%s text:%s}", c.addr.String(), c.text)
		}
		return fmt.Sprintf("{c command text:%s}", c.text)
	}
	return fmt.Sprintf("{c command}")
}

func (c *c_cmd) printText(s *Sed) {
	// we are going to get the newline from the
	s.outputFile.Write(c.text)
}

func (c *c_cmd) processLine(s *Sed) (bool, error) {
	s.patternSpace = s.patternSpace[0:0]
	if c.addr != nil {
		switch c.addr.address_type {
		case ADDRESS_RANGE:
			if s.lineNumber+1 == c.addr.rangeEnd {
				c.printText(s)
				return true, nil
			}
		case ADDRESS_LINE, ADDRESS_REGEX, ADDRESS_LAST_LINE:
			c.printText(s)
			return true, nil
		case ADDRESS_TO_END_OF_FILE:
			// FIX need to output at end of file
		}
	} else {
		c.printText(s)
		return true, nil
	}
	return false, nil
}

func NewCCmd(s *Sed, line []byte, addr *address) (*c_cmd, error) {
	cmd := new(c_cmd)
	cmd.addr = addr
	cmd.text = line[1:]
	for bytes.HasSuffix(cmd.text, []byte{'\\'}) {
		cmd.text = cmd.text[0 : len(cmd.text)-1]
		line, err := s.getNextScriptLine()
		if err != nil {
			break
		}
		buf := bytes.NewBuffer(cmd.text)
		buf.WriteRune('\n')
		buf.Write(line)
		cmd.text = buf.Bytes()
	}
	return cmd, nil
}
// E-OF: C_CMD //
// D_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)d%20Delete%20the%20pattern%20space.,newline.%20%20Start%20the%20next%20cycle. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)d%20Delete%20the%20pattern%20space.,newline.%20%20Start%20the%20next%20cycle.
type d_cmd struct {
	addr             *address
	upToFirstNewLine bool
}

func (c *d_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *d_cmd) String() string {
	if c != nil && c.addr != nil {
		return fmt.Sprintf("{d command addr:%s}", c.addr.String())
	}
	return fmt.Sprintf("{d command}")
}

func (c *d_cmd) processLine(s *Sed) (bool, error) {
	if c.upToFirstNewLine {
		idx := bytes.IndexByte(s.patternSpace, '\n')
		if idx >= 0 && idx+1 < len(s.patternSpace) {
			s.patternSpace = s.patternSpace[idx+1:]
			return false, nil
		}
	}
	return true, nil
}

func NewDCmd(pieces [][]byte, addr *address) (*d_cmd, error) {
	if len(pieces) > 1 {
		return nil, WrongNumberOfCommandParameters
	}
	cmd := new(d_cmd)
	if pieces[0][0] == 'D' {
		cmd.upToFirstNewLine = true
	}
	cmd.addr = addr
	return cmd, nil
}
// E-OF: D_CMD //
// EQL_CMD //
type eql_cmd struct {
	addr *address
}

func (c *eql_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *eql_cmd) String() string {
	if c != nil && c.addr != nil {
		return fmt.Sprintf("{= command addr: %s}", c.addr.String())
	}
	return fmt.Sprint("{= command}")
}

func (c *eql_cmd) processLine(s *Sed) (bool, error) {
	fmt.Fprintf(os.Stdout, "\n%d\n", s.lineNumber)
	return false, nil
}

func NewEqlCmd(pieces [][]byte, addr *address) (*eql_cmd, error) {
	if len(pieces) > 1 {
		return nil, WrongNumberOfCommandParameters
	}
	cmd := new(eql_cmd)
	cmd.addr = addr
	return cmd, nil
}
// E-OF: EQL_CMD //
// G_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)g%20Replace%20the%20contents%20of,tents%20of%20the%20hold%20space. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)g%20Replace%20the%20contents%20of,tents%20of%20the%20hold%20space.
type g_cmd struct {
	addr    *address
	replace bool
}

func (c *g_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *g_cmd) String() string {
	if c != nil {
		if c.addr != nil {
			if c.replace {
				return fmt.Sprintf("{g command with replace addr:%s}", c.addr.String())
			} else {
				return fmt.Sprintf("{g command addr:%s}", c.addr.String())
			}
		} else {
			if c.replace {
				return fmt.Sprint("{g command with replace}")
			} else {
				return fmt.Sprint("{Append a newline and the hold space to the pattern space}")
			}
		}
	}
	return fmt.Sprint("{Append/Replace pattern space with contents of hold space}")
}

func (c *g_cmd) processLine(s *Sed) (bool, error) {
	if c.replace {
		s.patternSpace = copyByteSlice(s.holdSpace)
	} else {
		buf := bytes.NewBuffer(s.patternSpace)
		buf.WriteRune('\n')
		buf.Write(s.holdSpace)
		s.patternSpace = buf.Bytes()
	}
	return false, nil
}

func NewGCmd(pieces [][]byte, addr *address) (*g_cmd, error) {
	if len(pieces) > 1 {
		return nil, WrongNumberOfCommandParameters
	}
	cmd := new(g_cmd)
	if pieces[0][0] == 'g' {
		cmd.replace = true
	}
	cmd.addr = addr
	return cmd, nil
}
// E-OF: G_CMD //
// H_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)h%20Replace%20the%20contents%20of,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20of%20the%20pattern%20space. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)h%20Replace%20the%20contents%20of,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20of%20the%20pattern%20space.
type h_cmd struct {
	addr    *address
	replace bool
}

func (c *h_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *h_cmd) String() string {
	if c != nil {
		if c.addr != nil {
			if c.replace {
				return fmt.Sprintf("{h command with replace addr:%s}", c.addr.String())
			} else {
				return fmt.Sprintf("{h command Cmd addr:%s}", c.addr.String())
			}
		} else {
			if c.replace {
				return fmt.Sprint("{h command with replace }")
			} else {
				return fmt.Sprint("{h command")
			}
		}
	}
	return fmt.Sprint("{h command}")
}

func (c *h_cmd) processLine(s *Sed) (bool, error) {
	if c.replace {
		s.holdSpace = copyByteSlice(s.patternSpace)
	} else {
		buf := bytes.NewBuffer(s.patternSpace)
		buf.WriteRune('\n')
		buf.Write(s.holdSpace)
		s.patternSpace = buf.Bytes()
	}
	return false, nil
}

func NewHCmd(pieces [][]byte, addr *address) (*h_cmd, error) {
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
// E-OF: H_CMD //
// I_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(1)i%5C,the%20standard%20output. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(1)i%5C,the%20standard%20output.
type i_cmd struct {
	addr *address
	text []byte
}

func (c *i_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *i_cmd) String() string {
	if c != nil {
		if c.addr != nil {
			return fmt.Sprintf("{i command addr:%s text:%s}", c.addr.String(), string(c.text))
		}
		return fmt.Sprintf("{i command text:%s}", string(c.text))
	}
	return fmt.Sprintf("{i command}")
}

func (c *i_cmd) processLine(s *Sed) (bool, error) {
	return false, nil
}

func NewICmd(s *Sed, line []byte, addr *address) (*i_cmd, error) {
	cmd := new(i_cmd)
	cmd.addr = addr
	cmd.text = line[1:]
	for bytes.HasSuffix(cmd.text, []byte{'\\'}) {
		cmd.text = cmd.text[0 : len(cmd.text)-1]
		line, err := s.getNextScriptLine()
		if err != nil {
			break
		}
		// cmd.text = bytes.AddByte(cmd.text, '\n')
		buf := bytes.NewBuffer(cmd.text)
		buf.Write(line)
		s.patternSpace = buf.Bytes()
	}
	return cmd, nil
}
// E-OF: I_CMD //
// N_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)n%20Copy%20the%20pattern%20space,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20changes.) // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)n%20Copy%20the%20pattern%20space,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20changes.)
type n_cmd struct {
	addr   *address
	append bool // Distinguishes between 'n' (false) and 'N' (true)
}

func (c *n_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *n_cmd) String() string {
	if c != nil && c.addr != nil {
		if c.append {
			return fmt.Sprintf("{N command addr:%s}", c.addr.String())
		}
		return fmt.Sprintf("{n command addr:%s}", c.addr.String())
	}
	return fmt.Sprint("{n/N command}")
}

func (c *n_cmd) processLine(s *Sed) (bool, error) {
	if c.append {
		// N: Append the next line of input to the pattern space
		nextLine, err := s.input.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return false, nil
			}
			return false, err
		}
		s.patternSpace = append(s.patternSpace, newLine...)
		s.patternSpace = append(s.patternSpace, nextLine...)
	} else {
		// n: Print and replace pattern space with the next line
		if !*quiet {
			s.printPatternSpace()
		}
		nextLine, err := s.input.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return false, nil
			}
			return false, err
		}
		s.patternSpace = nextLine
	}
	return true, nil
}

func NewNCmd(pieces [][]byte, addr *address) (*n_cmd, error) {
	if len(pieces) > 1 {
		return nil, WrongNumberOfCommandParameters
	}
	cmd := &n_cmd{
		addr:   addr,
		append: pieces[0][0] == 'N',
	}
	return cmd, nil
}
// E-OF: N_CMD //
// P_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)p%20Print.%20%20Copy%20the%20pattern,newline%20to%20the%20standard%20output. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)p%20Print.%20%20Copy%20the%20pattern,newline%20to%20the%20standard%20output.
type p_cmd struct {
	addr        *address
	upToNewLine bool
}

func (c *p_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *p_cmd) String() string {
	if c != nil && c.addr != nil {
		return fmt.Sprintf("{p command addr:%s}", c.addr.String())
	}
	return fmt.Sprint("{p command}")
}

func (c *p_cmd) processLine(s *Sed) (bool, error) {
	// print output space
	if c.upToNewLine {
		firstLine := bytes.SplitN(s.patternSpace, []byte{'\n'}, 1)[0]
		fmt.Fprintln(s.outputFile, string(firstLine))
	} else {
		fmt.Fprintln(s.outputFile, string(s.patternSpace))
	}
	return false, nil
}

func NewPCmd(pieces [][]byte, addr *address) (*p_cmd, error) {
	if len(pieces) > 1 {
		return nil, WrongNumberOfCommandParameters
	}
	cmd := new(p_cmd)
	cmd.addr = addr
	cmd.upToNewLine = pieces[0][0] == 'P'
	return cmd, nil
}
// E-OF: P_CMD //
// Q_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(1)q%20Quit.%20%20Branch%20to%20the,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20new%20cycle. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(1)q%20Quit.%20%20Branch%20to%20the,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20new%20cycle.
type q_cmd struct {
	addr      *address
	exit_code int
}

func (c *q_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *q_cmd) String() string {
	if c != nil {
		if c.addr != nil {
			return fmt.Sprintf("{q command addr:%s with exit code: %d}", c.addr.String(), c.exit_code)
		}
		return fmt.Sprintf("{q command with exit code: %d}", c.exit_code)
	}
	return fmt.Sprint("{q command}")
}

func NewQCmd(pieces [][]byte, addr *address) (c *q_cmd, err error) {
	err = nil
	c = nil
	switch len(pieces) {
	case 2:
		c = new(q_cmd)
		c.addr = addr
		c.exit_code, err = strconv.Atoi(string(pieces[1]))
		if err != nil {
			c = nil
		}
	case 1:
		c = new(q_cmd)
		c.addr = addr
		c.exit_code = 0
	default:
		c, err = nil, WrongNumberOfCommandParameters
	}
	return c, err
}

func (c *q_cmd) processLine(s *Sed) (stop bool, err error) {
	os.Exit(c.exit_code)
	return false, nil
}
// E-OF: Q_CMD //
// R_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)r%20rfile,reading%20the%20next%20input%20line. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)r%20rfile,reading%20the%20next%20input%20line.
type r_cmd struct {
	addr *address
	text []byte
}

func (c *r_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *r_cmd) String() string {
	if c != nil && c.addr != nil {
		return fmt.Sprintf("{r command addr:%s}", c.addr.String())
	}
	return fmt.Sprint("{r command}")
}

func (c *r_cmd) processLine(s *Sed) (bool, error) {
	// print output space
	if c.text != nil {
		s.outputFile.Write(c.text)
	}
	return false, nil
}

func NewRCmd(line []byte, addr *address) (*r_cmd, error) {
	line = line[1:]
	cmd := new(r_cmd)
	cmd.addr = addr
	if len(line) > 0 {
		cmd.text = line
	} else {
		cmd.text = nil
	}
	return cmd, nil
}
// E-OF: R_CMD //
// S_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)s/regular%2Dexpression/replacement/flags,regular%2Dexpression%20in%20the%20pattern%20space. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)s/regular%2Dexpression/replacement/flags,regular%2Dexpression%20in%20the%20pattern%20space.
const (
	global_replace = -1
)

type s_cmd struct {
	addr         *address
	regex        string
	replace      []byte
	nthOccurance int
	re           *regexp.Regexp
}

func (c *s_cmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *s_cmd) String() string {
	if c != nil {
		if c.addr != nil {
			return fmt.Sprintf("{s command addr:%s regex:%v replace:%s nth occurance:%d}", c.addr, c.regex, c.replace, c.nthOccurance)
		}
		return fmt.Sprintf("{s command regex:%v replace:%s nth occurance:%d}", c.regex, c.replace, c.nthOccurance)
	}
	return "{s command}"
}

func NewSCmd(pieces [][]byte, addr *address) (c *s_cmd, err error) {
	if len(pieces) != 4 {
		return nil, WrongNumberOfCommandParameters
	}

	err = nil
	c = new(s_cmd)
	c.addr = addr

	c.regex = string(pieces[1])
	if len(c.regex) == 0 {
		return nil, RegularExpressionExpected
	}
	c.re, err = regexp.CompilePOSIX(string(c.regex))
	if err != nil {
		return nil, err
	}

	c.replace = pieces[2]

	flag := string(pieces[3])
	if flag != "g" {
		c.nthOccurance = 1
		if len(flag) > 0 {
			c.nthOccurance, err = strconv.Atoi(flag)
			if err != nil {
				return nil, InvalidSCommandFlag
			}
		}
	} else {
		c.nthOccurance = global_replace
	}

	return c, err
}

func (c *s_cmd) processLine(s *Sed) (stop bool, err error) {
	stop, err = false, nil

	switch c.nthOccurance {
	case global_replace:
		s.patternSpace = c.re.ReplaceAll(s.patternSpace, c.replace)
	default:
		// a numeric flag command
		count := 0
		line := s.patternSpace
		s.patternSpace = make([]byte, 0)
		for {
			matches := c.re.FindIndex(line)
			if len(matches) > 0 {
				count++
				if count == c.nthOccurance {
					buf := bytes.NewBuffer(s.patternSpace)
					buf.Write(line[0:matches[0]])
					buf.Write(c.replace)
					buf.Write(line[matches[1]:])
					s.patternSpace = buf.Bytes()
					break
				} else {
					buf := bytes.NewBuffer(s.patternSpace)
					buf.Write(line[0 : matches[0]+1])
					s.patternSpace = buf.Bytes()
				}
				line = line[matches[0]+1:]
			} else {
				buf := bytes.NewBuffer(s.patternSpace)
				buf.Write(line)
				s.patternSpace = buf.Bytes()
				break
			}
		}
	}
	return stop, err
}
// E-OF: S_CMD //
