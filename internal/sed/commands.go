// ./definitions/*
// sed
//
// Original code: Copyright (c) 2009 Geoffrey Clements (MIT License)
// Modified code: Copyright (c) 2024 xplshn (3BSD License)
// For details, see LICENSE file in the root directory of this project.

// In order to generate commands.go, you must do `cat ./definitions/*.gopart > ./commands.go`
// This is why a_cmd.gopart is the only file that includes the License header

// A_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(1)a%5C,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20next%20input%20line. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(1)a%5C,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20next%20input%20line.
package sed

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
)

// Used in other parts of the `sed` package.
const (
	globalReplace = -1
)


// ACmd represents an 'a' command in sed.
type ACmd struct {
	addr *address
	text []byte
}

func (c *ACmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

func (c *ACmd) String() string {
	if c != nil {
		if c.addr != nil {
			return fmt.Sprintf("{a command addr:%s text:%s}", c.addr.String(), c.text)
		}
		return fmt.Sprintf("{a command text:%s}", c.text)
	}
	return fmt.Sprintf("{a command}")
}

func (c *ACmd) processLine(_ *Sed) (bool, error) {
	return false, nil
}

// NewACmd creates a new aCmd instance from the given Sed object, line, and address.
func NewACmd(s *Sed, line []byte, addr *address) (*ACmd, error) {
	cmd := new(ACmd)
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

// BCmd represents a 'b' command in sed, which branches to a specified label
type BCmd struct {
	addr  *address
	label string
}

// match checks if the given line matches the address criteria of the bCmd.
func (c *BCmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

// String returns a string representation of the BCmd, including its label and address
func (c *BCmd) String() string {
	if c != nil {
		if c.addr != nil {
			return fmt.Sprintf("{b command label: %s Cmd addr:%s}", c.label, c.addr.String())
		}
		return fmt.Sprintf("{b command label: %s Cmd}", c.label)
	}
	return fmt.Sprintf("{b command}")
}

// processLine processes the input line for the BCmd. It returns an error indicating the function is not implemented.
func (c *BCmd) processLine(_ *Sed) (bool, error) {
	return true, ErrNotImplemented
}
// NewBCmd creates a new BCmd instance from the given pieces of input and address
func NewBCmd(pieces [][]byte, addr *address) (*BCmd, error) {
	if len(pieces) != 1 {
		return nil, ErrWrongNumberOfCommandParameters
	}
	cmd := new(BCmd)
	cmd.addr = addr
	cmd.label = string(bytes.TrimSpace(pieces[0][1:]))
	return cmd, nil
}

// E-OF: B_CMD //
// C_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)c%5C,output.%20%20Start%20the%20next%20cycle. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)c%5C,output.%20%20Start%20the%20next%20cycle.

// CCmd represents a 'c' command in sed, which replaces lines that match the address with specified text.
type CCmd struct {
	addr *address
	text []byte
}

// match checks if the given line matches the address criteria of the CCmd.
func (c *CCmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

// String returns a string representation of the CCmd, including its address and text.
func (c *CCmd) String() string {
	if c.addr != nil {
		return fmt.Sprintf("{c command addr:%s text:%s}", c.addr.String(), string(c.text))
	}
	return fmt.Sprintf("{c command text:%s}", string(c.text))
}

// printText writes the command's text to the output file.
func (c *CCmd) printText(s *Sed) {
	_, err := s.outputFile.Write(c.text)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing text: %v\n", err)
	}
}

// processLine processes the input line for the CCmd, replacing the content based on the address.
func (c *CCmd) processLine(s *Sed) (bool, error) {
	s.patternSpace = s.patternSpace[:0]
	if c.addr != nil {
		switch c.addr.addressType {
		case addressRange:
			if s.lineNumber+1 == c.addr.rangeEnd {
				c.printText(s)
				return true, nil
			}
		case addressLine, addressRegEx, addressLastLine:
			c.printText(s)
			return true, nil
		case addressToEndOfFile:
			// Output at end of file is not handled here
			fmt.Fprintln(os.Stderr, "TODO: Handle output at end of file")
		}
	} else {
		c.printText(s)
		return true, nil
	}
	return false, nil
}

// NewCCmd creates a new CCmd instance from the given Sed object, line of input, and address.
func NewCCmd(s *Sed, line []byte, addr *address) (*CCmd, error) {
	cmd := &CCmd{
		addr: addr,
		text: line[1:],
	}
	for bytes.HasSuffix(cmd.text, []byte{'\\'}) {
		cmd.text = cmd.text[:len(cmd.text)-1]
		nextLine, err := s.getNextScriptLine()
		if err != nil {
			return nil, err
		}
		buf := bytes.NewBuffer(cmd.text)
		buf.WriteRune('\n')
		buf.Write(nextLine)
		cmd.text = buf.Bytes()
	}
	return cmd, nil
}

// E-OF: C_CMD //
// D_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)d%20Delete%20the%20pattern%20space.,newline.%20%20Start%20the%20next%20cycle. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)d%20Delete%20the%20pattern%20space.,newline.%20%20Start%20the%20next%20cycle.

// DCmd represents a 'd' command in sed, which deletes the pattern space up to the first newline or entirely.
type DCmd struct {
	addr             *address
	upToFirstNewLine bool
}

// match checks if the given line matches the address criteria of the DCmd.
func (c *DCmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

// String returns a string representation of the DCmd, including its address and whether it deletes up to the first newline.
func (c *DCmd) String() string {
	if c.addr != nil {
		if c.upToFirstNewLine {
			return fmt.Sprintf("{d command addr:%s up to first newline}", c.addr.String())
		}
		return fmt.Sprintf("{d command addr:%s}", c.addr.String())
	}
	if c.upToFirstNewLine {
		return "{d command up to first newline}"
	}
	return "{d command}"
}

// processLine processes the input line for the DCmd, deleting the pattern space up to the first newline if specified.
func (c *DCmd) processLine(s *Sed) (bool, error) {
	if c.upToFirstNewLine {
		idx := bytes.IndexByte(s.patternSpace, '\n')
		if idx >= 0 && idx+1 < len(s.patternSpace) {
			s.patternSpace = s.patternSpace[idx+1:]
		} else {
			s.patternSpace = s.patternSpace[:0] // Clear pattern space if newline is not found
		}
	}
	return true, nil
}

// NewDCmd creates a new DCmd instance from the given pieces of input and address.
func NewDCmd(pieces [][]byte, addr *address) (*DCmd, error) {
	if len(pieces) > 1 {
		return nil, ErrWrongNumberOfCommandParameters
	}
	cmd := &DCmd{
		addr: addr,
	}
	if len(pieces) > 0 && pieces[0][0] == 'D' {
		cmd.upToFirstNewLine = true
	}
	return cmd, nil
}

// E-OF: dCmd //
// EQL_CMD //

// EqlCmd represents an '=' command in sed, which prints the current line number.
type EqlCmd struct {
	addr *address
}

// match checks if the given line matches the address criteria of the EqlCmd.
func (c *EqlCmd) match(line []byte, lineNumber int) bool {
    return c.addr.match(line, lineNumber)
}

// String returns a string representation of the EqlCmd, including its address.
func (c *EqlCmd) String() string {
    if c != nil && c.addr != nil {
        return fmt.Sprintf("{= command addr: %s}", c.addr.String())
    }
    return fmt.Sprint("{= command}")
}

// processLine processes the input line for the EqlCmd, printing the current line number.
func (c *EqlCmd) processLine(s *Sed) (bool, error) {
    fmt.Fprintf(os.Stdout, "\n%d\n", s.lineNumber)
    return false, nil
}

// NewEqlCmd creates a new EqlCmd instance from the given pieces of input and address.
func NewEqlCmd(pieces [][]byte, addr *address) (*EqlCmd, error) {
    if len(pieces) > 1 {
        return nil, ErrWrongNumberOfCommandParameters
    }
    cmd := new(EqlCmd)
    cmd.addr = addr
    return cmd, nil
}

// E-OF: EQL_CMD //
// G_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)g%20Replace%20the%20contents%20of,tents%20of%20the%20hold%20space. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)g%20Replace%20the%20contents%20of,tents%20of%20the%20hold%20space.

// GCmd represents a 'g' command in sed, which replaces or appends the contents of the hold space to the pattern space.
type GCmd struct {
	addr    *address
	replace bool
}

// match checks if the given line matches the address criteria of the GCmd.
func (c *GCmd) match(line []byte, lineNumber int) bool {
    return c.addr.match(line, lineNumber)
}

// String returns a string representation of the GCmd, including its address and replace status.
func (c *GCmd) String() string {
    if c != nil {
        if c.addr != nil {
            if c.replace {
                return fmt.Sprintf("{g command with replace addr:%s}", c.addr.String())
            }
            return fmt.Sprintf("{g command addr:%s}", c.addr.String())
        }
        if c.replace {
            return fmt.Sprint("{g command with replace}")
        }
        return fmt.Sprint("{Append a newline and the hold space to the pattern space}")
    }
    return fmt.Sprint("{Append/Replace pattern space with contents of hold space}")
}

// processLine processes the input line for the GCmd, replacing or appending the hold space as specified.
func (c *GCmd) processLine(s *Sed) (bool, error) {
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

// NewGCmd creates a new GCmd instance from the given pieces of input and address.
func NewGCmd(pieces [][]byte, addr *address) (*GCmd, error) {
    if len(pieces) > 1 {
        return nil, ErrWrongNumberOfCommandParameters
    }
    cmd := new(GCmd)
    if pieces[0][0] == 'g' {
        cmd.replace = true
    }
    cmd.addr = addr
    return cmd, nil
}

// E-OF: G_CMD //
// H_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)h%20Replace%20the%20contents%20of,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20of%20the%20pattern%20space. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)h%20Replace%20the%20contents%20of,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20of%20the%20pattern%20space.

// HCmd represents an 'h' command in sed, which replaces or appends the contents of the pattern space to the hold space.
type HCmd struct {
	addr    *address
	replace bool
}

// match checks if the given line matches the address criteria of the HCmd.
func (c *HCmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

// String returns a string representation of the HCmd, including its address and replace status.
func (c *HCmd) String() string {
	if c != nil {
		if c.addr != nil {
			if c.replace {
				return fmt.Sprintf("{h command with replace addr:%s}", c.addr.String())
			}
			return fmt.Sprintf("{h command Cmd addr:%s}", c.addr.String())
		}
		if c.replace {
			return fmt.Sprint("{h command with replace }")
		}
		return fmt.Sprint("{h command")
	}
	return fmt.Sprint("{h command}")
}

// processLine processes the input line for the HCmd, replacing or appending the pattern space as specified.
func (c *HCmd) processLine(s *Sed) (bool, error) {
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

// NewHCmd creates a new HCmd instance from the given pieces of input and address.
func NewHCmd(pieces [][]byte, addr *address) (*HCmd, error) {
	if len(pieces) > 1 {
		return nil, ErrWrongNumberOfCommandParameters
	}
	cmd := new(HCmd)
	if pieces[0][0] == 'h' {
		cmd.replace = true
	}
	cmd.addr = addr
	return cmd, nil
}

// E-OF: H_CMD //
// I_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(1)i%5C,the%20standard%20output. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(1)i%5C,the%20standard%20output.

// ICmd represents an 'i' command in sed, which inserts text before the current pattern space and outputs it to the standard output.
type ICmd struct {
	addr *address
	text []byte
}

// match checks if the given line matches the address criteria of the ICmd.
func (c *ICmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

// String returns a string representation of the ICmd, including its address and text.
func (c *ICmd) String() string {
	if c != nil {
		if c.addr != nil {
			return fmt.Sprintf("{i command addr:%s text:%s}", c.addr.String(), string(c.text))
		}
		return fmt.Sprintf("{i command text:%s}", string(c.text))
	}
	return fmt.Sprintf("{i command}")
}

// processLine processes the input line for the ICmd. It does not alter the pattern space.
func (c *ICmd) processLine(s *Sed) (bool, error) {
	s.patternSpace = append([]byte{}, c.text...)
	s.patternSpace = append(s.patternSpace, '\n')
	return false, nil
}

// NewICmd creates a new ICmd instance from the given line of input and address.
func NewICmd(s *Sed, line []byte, addr *address) (*ICmd, error) {
	cmd := new(ICmd)
	cmd.addr = addr
	cmd.text = line[1:]
	for bytes.HasSuffix(cmd.text, []byte{'\\'}) {
		cmd.text = cmd.text[0 : len(cmd.text)-1]
		line, err := s.getNextScriptLine()
		if err != nil {
			break
		}
		buf := bytes.NewBuffer(cmd.text)
		buf.Write(line)
		s.patternSpace = buf.Bytes()
	}
	return cmd, nil
}

// E-OF: I_CMD //
// N_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)n%20Copy%20the%20pattern%20space,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20changes.) // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)n%20Copy%20the%20pattern%20space,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20changes.)

// NCmd represents an 'n' command in sed, which either prints the pattern space and replaces it with the next line ('n') or appends the next line of input to the pattern space ('N').
type NCmd struct {
	addr   *address
	append bool // Distinguishes between 'n' (false) and 'N' (true)
}

// match checks if the given line matches the address criteria of the NCmd.
func (c *NCmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

// String returns a string representation of the NCmd, including its address and whether it appends or replaces.
func (c *NCmd) String() string {
	if c != nil {
		if c.addr != nil {
			if c.append {
				return fmt.Sprintf("{N command addr:%s}", c.addr.String())
			}
			return fmt.Sprintf("{n command addr:%s}", c.addr.String())
		}
		if c.append {
			return fmt.Sprint("{N command}")
		}
		return fmt.Sprint("{n command}")
	}
	return fmt.Sprint("{n/N command}")
}

// processLine processes the input line for the NCmd. It either prints the pattern space and replaces it with the next line or appends the next line to the pattern space.
func (c *NCmd) processLine(s *Sed) (bool, error) {
	if c.append {
		// N: Append the next line of input to the pattern space
		nextLine, err := s.input.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return false, nil
			}
			return false, err
		}
		s.patternSpace = append(s.patternSpace, '\n')
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

// NewNCmd creates a new NCmd instance from the given pieces and address.
func NewNCmd(pieces [][]byte, addr *address) (*NCmd, error) {
	if len(pieces) > 1 {
		return nil, ErrWrongNumberOfCommandParameters
	}
	cmd := &NCmd{
		addr:   addr,
		append: pieces[0][0] == 'N',
	}
	return cmd, nil
}

// E-OF: N_CMD //
// P_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)p%20Print.%20%20Copy%20the%20pattern,newline%20to%20the%20standard%20output. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)p%20Print.%20%20Copy%20the%20pattern,newline%20to%20the%20standard%20output.

// PCmd represents a 'p' command in sed, which prints the pattern space. It can be configured to print up to the first newline or the entire pattern space.
type PCmd struct {
	addr        *address
	upToNewLine bool // If true, prints only up to the first newline; otherwise, prints the entire pattern space
}

// match checks if the given line matches the address criteria of the PCmd.
func (c *PCmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

// String returns a string representation of the PCmd, including its address and whether it prints up to a newline.
func (c *PCmd) String() string {
	if c != nil {
		if c.addr != nil {
			if c.upToNewLine {
				return fmt.Sprintf("{p command addr:%s up to newline}", c.addr.String())
			}
			return fmt.Sprintf("{p command addr:%s}", c.addr.String())
		}
		if c.upToNewLine {
			return fmt.Sprint("{p command up to newline}")
		}
		return fmt.Sprint("{p command}")
	}
	return fmt.Sprint("{p command}")
}

// processLine processes the pattern space for the PCmd. It either prints up to the first newline or the entire pattern space.
func (c *PCmd) processLine(s *Sed) (bool, error) {
	if c.upToNewLine {
		// Print only up to the first newline
		firstLine := bytes.SplitN(s.patternSpace, []byte{'\n'}, 2)[0]
		fmt.Fprintln(s.outputFile, string(firstLine))
	} else {
		// Print the entire pattern space
		fmt.Fprintln(s.outputFile, string(s.patternSpace))
	}
	return false, nil
}

// NewPCmd creates a new PCmd instance from the given pieces and address.
func NewPCmd(pieces [][]byte, addr *address) (*PCmd, error) {
	if len(pieces) > 1 {
		return nil, ErrWrongNumberOfCommandParameters
	}
	cmd := &PCmd{
		addr:        addr,
		upToNewLine: pieces[0][0] == 'P',
	}
	return cmd, nil
}

// E-OF: P_CMD //
// Q_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(1)q%20Quit.%20%20Branch%20to%20the,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20new%20cycle. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(1)q%20Quit.%20%20Branch%20to%20the,%20%20%20%20%20%20%20%20%20%20%20%20%20%20%20new%20cycle.

// QCmd represents a 'q' command in sed, which terminates the sed process with a specified exit code.
type QCmd struct {
	addr     *address
	exitCode int // The exit code to return when the 'q' command is executed
}

// match checks if the given line matches the address criteria of the QCmd.
func (c *QCmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

// String returns a string representation of the QCmd, including its address and exit code.
func (c *QCmd) String() string {
	if c != nil {
		if c.addr != nil {
			return fmt.Sprintf("{q command addr:%s with exit code: %d}", c.addr.String(), c.exitCode)
		}
		return fmt.Sprintf("{q command with exit code: %d}", c.exitCode)
	}
	return fmt.Sprint("{q command}")
}

// NewQCmd creates a new QCmd instance from the given pieces and address.
// It parses the exit code if provided, or defaults to 0.
func NewQCmd(pieces [][]byte, addr *address) (*QCmd, error) {
	var err error
	cmd := &QCmd{
		addr: addr,
	}
	switch len(pieces) {
	case 2:
		cmd.exitCode, err = strconv.Atoi(string(pieces[1]))
		if err != nil {
			return nil, err
		}
	case 1:
		cmd.exitCode = 0
	default:
		return nil, ErrWrongNumberOfCommandParameters
	}
	return cmd, nil
}

// processLine terminates the sed process with the exit code specified in the QCmd.
func (c *QCmd) processLine(_ *Sed) (bool, error) {
	os.Exit(c.exitCode)
	return false, nil
}

// E-OF: Q_CMD //
// R_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)r%20rfile,reading%20the%20next%20input%20line. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)r%20rfile,reading%20the%20next%20input%20line.

// RCmd represents an 'r' command in sed, which reads a file and appends its contents to the pattern space.
type RCmd struct {
	addr *address
	text []byte // Text to be written to the output file
}

// match checks if the given line matches the address criteria of the RCmd.
func (c *RCmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

// String returns a string representation of the RCmd, including its address and the text.
func (c *RCmd) String() string {
	if c.addr != nil {
		return fmt.Sprintf("{r command addr:%s}", c.addr.String())
	}
	return fmt.Sprint("{r command}")
}

// processLine writes the stored text to the output file if it exists.
func (c *RCmd) processLine(s *Sed) (bool, error) {
	if c.text != nil {
		_, err := s.outputFile.Write(c.text)
		if err != nil {
			return false, err
		}
	}
	return false, nil
}

// NewRCmd creates a new RCmd instance from the given line and address.
// It initializes the command with the text specified after the 'r' command, if provided.
func NewRCmd(line []byte, addr *address) (*RCmd, error) {
	if len(line) > 1 {
		line = line[1:] // Remove the initial 'r' character
	} else {
		line = nil
	}
	cmd := &RCmd{
		addr: addr,
		text: line,
	}
	return cmd, nil
}

// E-OF: R_CMD //
// S_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)s/regular%2Dexpression/replacement/flags,regular%2Dexpression%20in%20the%20pattern%20space. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)s/regular%2Dexpression/replacement/flags,regular%2Dexpression%20in%20the%20pattern%20space.

// SCmd represents an 's' command in sed, which performs a substitution based on a regular expression in the pattern space.
type SCmd struct {
	addr         *address
	regex        string
	replace      []byte
	nthOccurance int
	re           *regexp.Regexp
}

// match checks if the given line matches the address criteria of the SCmd.
func (c *SCmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

// String returns a string representation of the SCmd, including its address, regex, replacement, and nth occurrence.
func (c *SCmd) String() string {
	if c.addr != nil {
		return fmt.Sprintf("{s command addr:%s regex:%v replace:%s nth occurrence:%d}", c.addr, c.regex, c.replace, c.nthOccurance)
	}
	return fmt.Sprintf("{s command regex:%v replace:%s nth occurrence:%d}", c.regex, c.replace, c.nthOccurance)
}

// NewSCmd creates a new SCmd instance from the given pieces of input and address.
func NewSCmd(pieces [][]byte, addr *address) (*SCmd, error) {
	if len(pieces) != 4 {
		return nil, ErrWrongNumberOfCommandParameters
	}

	cmd := &SCmd{
		addr: addr,
		regex: string(pieces[1]),
		replace: pieces[2],
	}

	if len(cmd.regex) == 0 {
		return nil, ErrRegularExpressionExpected
	}

	var err error
	cmd.re, err = regexp.CompilePOSIX(cmd.regex)
	if err != nil {
		return nil, err
	}

	flag := string(pieces[3])
	if flag == "g" {
		cmd.nthOccurance = globalReplace
	} else {
		cmd.nthOccurance = 1
		if len(flag) > 0 {
			cmd.nthOccurance, err = strconv.Atoi(flag)
			if err != nil {
				return nil, ErrInvalidSCommandFlag
			}
		}
	}

	return cmd, nil
}

// processLine processes the input line for the SCmd, performing substitutions based on the regular expression.
func (c *SCmd) processLine(s *Sed) (bool, error) {
	if c.nthOccurance == globalReplace {
		s.patternSpace = c.re.ReplaceAll(s.patternSpace, c.replace)
		return false, nil
	}

	count := 0
	line := s.patternSpace
	s.patternSpace = make([]byte, 0)
	for {
		matches := c.re.FindIndex(line)
		if len(matches) > 0 {
			count++
			if count == c.nthOccurance {
				buf := bytes.NewBuffer(s.patternSpace)
				buf.Write(line[:matches[0]])
				buf.Write(c.replace)
				buf.Write(line[matches[1]:])
				s.patternSpace = buf.Bytes()
				break
			}
			buf := bytes.NewBuffer(s.patternSpace)
			buf.Write(line[:matches[0]+1])
			s.patternSpace = buf.Bytes()
			line = line[matches[0]+1:]
			continue
		}
		buf := bytes.NewBuffer(s.patternSpace)
		buf.Write(line)
		s.patternSpace = buf.Bytes()
		break
	}
	return false, nil
}

// E-OF: S_CMD //
// X_CMD // As defined in: https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)x%20Exchange%20the%20contents%20of%20the%20pattern%20and%20hold%20spaces. // PERMALINK: https://web.archive.org/web/20240730163415/https://man.cat-v.org/unix_10th/1/sed#:~:text=(2)x%20Exchange%20the%20contents%20of%20the%20pattern%20and%20hold%20spaces.

// XCmd represents an 'x' command in sed, which exchanges the contents of the pattern and hold spaces.
type XCmd struct {
	addr *address
}

// match checks if the given line matches the address criteria of the XCmd.
func (c *XCmd) match(line []byte, lineNumber int) bool {
	return c.addr.match(line, lineNumber)
}

// String returns a string representation of the XCmd, including its address.
func (c *XCmd) String() string {
	if c != nil {
		if c.addr != nil {
			return fmt.Sprintf("{x command addr:%s}", c.addr.String())
		}
		return fmt.Sprintf("{x command}")
	}
	return fmt.Sprintf("{x command}")
}

// processLine processes the input line for the XCmd, exchanging the contents of the pattern and hold spaces.
func (c *XCmd) processLine(s *Sed) (bool, error) {
	// Exchange the contents of the pattern space and hold space
	s.patternSpace, s.holdSpace = s.holdSpace, s.patternSpace
	return false, nil
}

// NewXCmd creates a new XCmd instance from the given pieces of input and address.
func NewXCmd(pieces [][]byte, addr *address) (*XCmd, error) {
	if len(pieces) > 1 {
		return nil, ErrWrongNumberOfCommandParameters
	}
	cmd := new(XCmd)
	cmd.addr = addr
	return cmd, nil
}

// E-OF: X_CMD //
