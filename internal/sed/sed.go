// sed.go
// sed
//
// Original code: Copyright (c) 2009 Geoffrey Clements (MIT License)
// Modified code: Copyright (c) 2024 xplshn (3BSD License)
// For details, see the [LICENSE](https://github.com/xplshn/gosed) file at the root directory of this project
package sed

import (
	"bufio"
	"bytes"
	"container/list"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

const (
	versionMajor = 0
	versionMinor = 2
	versionPoint = 1
)

var versionString string
var quiet = flag.Bool("n", false, "Suppress automatic printing of pattern space.")
var script = flag.String("e", "", "Expression to process input. Can be provided as a string.")
var scriptFile = flag.String("f", "", "Read expression/script from a file. Ignored if -e is specified.")
var editInplace = flag.Bool("i", false, "Edit files in place. If not set, output is printed to stdout.")
var lineWrap = 0 // var lineWrap = flag.Uint("l", 0, "Specify the default line-wrap length for the l command. A length of 0 (zero) means to never wrap long lines. If not specified, it is taken to be 70.")
var usageShown = false
var newLine = []byte{'\n'}

func init() {
	versionString = fmt.Sprintf("%d.%d.%d", versionMajor, versionMinor, versionPoint)
}

// Sed holds the current file structure and operations
type Sed struct {
	inputFile               *os.File
	input                   *bufio.Reader
	lineNumber              int
	currentLine             string
	beforeCommands          *list.List
	commands                *list.List
	afterCommands           *list.List
	outputFile              *os.File
	patternSpace, holdSpace []byte
	scriptLines             [][]byte
	scriptLineNumber        int
}

// Init initializes the Sed instance by setting up the command lists and output file.
func (s *Sed) Init() {
	s.beforeCommands = new(list.List)
	s.commands = new(list.List)
	s.afterCommands = new(list.List)
	s.outputFile = os.Stdout
	s.patternSpace = make([]byte, 0)
	s.holdSpace = make([]byte, 0)
}

func copyByteSlice(a []byte) []byte {
	newSlice := make([]byte, len(a))
	copy(newSlice, a)
	return newSlice
}

var inputFilename string

func (s *Sed) getNextScriptLine() ([]byte, error) {
	if s.scriptLineNumber < len(s.scriptLines) {
		val := s.scriptLines[s.scriptLineNumber]
		s.scriptLineNumber++
		return val, nil
	}
	return nil, io.EOF
}

func trimSpaceFromBeginning(s []byte) []byte {
	start, end := 0, len(s)
	for start < end {
		r, wid := utf8.DecodeRune(s[start:end])
		if !unicode.IsSpace(r) {
			break
		}
		start += wid
	}
	return s[start:end]
}

func (s *Sed) parseScript(scriptBuffer []byte) (err error) {
	// Split the script buffer into lines
	s.scriptLines = bytes.Split(scriptBuffer, newLine)
	s.scriptLineNumber = 0

	for {
		line, err := s.getNextScriptLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Trim leading and trailing whitespace
		line = trimSpaceFromBeginning(line)
		if len(line) == 0 {
			// Skip empty lines
			continue
		}

		// Check if the line is a comment
		if line[0] == '#' {
			// Special case for -n flag
			if len(line) > 1 && line[1] == 'n' && s.scriptLineNumber == 1 {
				*quiet = true
			}
			continue
		}

		// Process the command
		c, err := NewCmd(s, line)
		if err != nil {
			fmt.Printf("Script error: %s -> %d: %s\n", err.Error(), s.scriptLineNumber, line)
			os.Exit(-1)
		}

		// Add the command to the appropriate list
		if _, ok := c.(*ICmd); ok {
			s.beforeCommands.PushBack(c)
		} else if _, ok := c.(*ACmd); ok {
			s.afterCommands.PushBack(c)
		} else {
			s.commands.PushBack(c)
		}
	}
	return nil
}

func (s *Sed) printLine(line []byte) {
	l := len(line)
	if lineWrap <= 0 || l < int(lineWrap) {
		fmt.Fprintf(s.outputFile, "%s\n", line)
	} else {
		// print the line in segments
		for i := 0; i < l; i += int(lineWrap) {
			endOfLine := i + int(lineWrap)
			if endOfLine > l {
				endOfLine = l
			}
			fmt.Fprintf(s.outputFile, "%s\n", line[i:endOfLine])
		}
	}
}

func (s *Sed) printPatternSpace() {
	lines := bytes.Split(s.patternSpace, newLine)
	for _, line := range lines {
		s.printLine(line)
	}
}

func (s *Sed) process() {
	if *editInplace {
		s.lineNumber = 0
	}
	var err error
	s.patternSpace, err = s.input.ReadSlice('\n')
	for err != io.EOF {
		lineLength := len(s.patternSpace)
		if lineLength > 0 {
			s.patternSpace = s.patternSpace[0 : lineLength-1]
		}
		s.currentLine = string(s.patternSpace)
		// track line number starting with line 1
		s.lineNumber++
		stop := false
		// process i commands
		for c := s.beforeCommands.Front(); c != nil; c = c.Next() {
			// ask the sed if we should process this command, based on address
			if cmd, ok := c.Value.(*ICmd); ok {
				if c.Value.(Address).match(s.patternSpace, s.lineNumber) {
					s.outputFile.Write(cmd.text)
				}
			}
		}
		for c := s.commands.Front(); c != nil; c = c.Next() {
			// ask the sed if we should process this command, based on address
			if c.Value.(Address).match(s.patternSpace, s.lineNumber) {
				var err error
				stop, err = c.Value.(Cmd).processLine(s)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
					fmt.Fprintf(os.Stderr, "Line: %d:%s\n", s.lineNumber, s.currentLine)
					fmt.Fprintf(os.Stderr, "Command: %s\n", c.Value.(Cmd).String())
					os.Exit(-1)
				}
				if stop {
					break
				}
			}
		}
		if !*quiet && !stop {
			s.printPatternSpace()
		}
		// process a commands
		for c := s.afterCommands.Front(); c != nil; c = c.Next() {
			// ask the sed if we should process this command, based on address
			if cmd, ok := c.Value.(*ACmd); ok {
				if c.Value.(Address).match(s.patternSpace, s.lineNumber) {
					fmt.Fprintf(s.outputFile, "%s\n", cmd.text)
				}
			}
		}
		s.patternSpace, err = s.input.ReadSlice('\n')
	}
}

// Main is the entrypoint of this program. The ../../main.go calls `sed.Main()` to get here and get things done.
func Main() {
	var err error
	s := new(Sed)
	s.Init()

	printHelpPage := func() {
		// only show and calculate usage once
		if !usageShown {
			cmdInfo := &ccmd.CmdInfo{
				Authors:     []string{"Geoffrey Clements", "xplshn"},
				Name:        "sed",
				Synopsis:    "[options] <script> <input_file>",
				Description: "Unix's standard Stream Editor",
				Notes:       "This version of sed is a redistribution with modifications of `https://github.com/baldmountain/gosed`",
				Since:       2009,
			}
			// Calculate/Generate the help page
			helpPage, err := cmdInfo.GenerateHelpPage()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error generating help page:", err)
				os.Exit(1)
			}
			fmt.Fprint(os.Stdout, (helpPage))
			usageShown = true
		}
	}
	flag.Usage = func() {
		printHelpPage()
	}
	flag.Parse()

	// The first parameter may be a script or an input file. This helps us track which
	currentFileParameter := 0
	var scriptBuffer []byte

	// We need a script
	if len(*script) == 0 {
		// No -e so try -f
		if len(*scriptFile) > 0 {
			sb, err := os.ReadFile(*scriptFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading script file %s\n", *scriptFile)
				os.Exit(-1)
			}
			scriptBuffer = sb
		} else if flag.NArg() > 0 { // Changed from > 1 to > 0 to correctly handle the case with a single argument as script
			script := flag.Arg(0)
			scriptBuffer = []byte(script)

			// Change semicolons to newlines for scripts on command line
			scriptBuffer = bytes.ReplaceAll(scriptBuffer, []byte(";"), []byte("\n"))

			// First parameter was the script, so move to the second parameter
			currentFileParameter++
		}
	} else {
		scriptBuffer = []byte(*script)
		// Change semicolons to newlines for scripts on command line
		scriptBuffer = bytes.ReplaceAll(scriptBuffer, []byte(";"), []byte("\n"))
	}

	// If script still isn't set, we are screwed, exit.
	if len(scriptBuffer) == 0 {
		printHelpPage()
		fmt.Fprint(os.Stderr, "error, no input script found.\n")
		os.Exit(-1)
	}

	// Parse script
	s.parseScript(scriptBuffer)

	if currentFileParameter >= flag.NArg() {
		if *editInplace {
			fmt.Fprintf(os.Stderr, "Warning: Option -i ignored\n")
		}
		s.input = bufio.NewReader(os.Stdin)
		s.process()
	} else {
		for ; currentFileParameter < flag.NArg(); currentFileParameter++ {
			inputFilename = flag.Arg(currentFileParameter)
			// actually do the processing
			s.inputFile, err = os.Open(inputFilename)
			if err != nil {
				printHelpPage()
				fmt.Fprintf(os.Stderr, "error, could not open input file: %s.\n", inputFilename)
				os.Exit(-1)
			}
			s.input = bufio.NewReader(s.inputFile)
			var tempFilename string
			if *editInplace {
				tempFilename = inputFilename + ".tmp"
				tmpc := 0
				dir, _ := os.Stat(tempFilename)
				for dir != nil {
					tmpc++
					tempFilename = inputFilename + "-" + strconv.Itoa(tmpc) + ".tmp"
					dir, _ = os.Stat(tempFilename)
				}
				f, err := os.Create(tempFilename)
				if err != nil {
					s.inputFile.Close()
					fmt.Fprintf(os.Stderr, "Error opening temp file file for inplace editing: %s\n", err.Error())
					os.Exit(-1)
				}
				s.outputFile = f
			}
			s.process()
			// done processing, close input file
			s.inputFile.Close()
			s.input = nil
			if *editInplace {
				s.outputFile.Seek(0, 0)
				// find out about
				dir, err := os.Stat(inputFilename)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error getting information about input file: %s %v\n", inputFilename, err)
					// os.Remove(tempFilename);
					os.Exit(-1)
				}
				// reopen input file
				s.inputFile, err = os.OpenFile(inputFilename, os.O_WRONLY|os.O_TRUNC, dir.Mode())
				if err != nil {
					fmt.Fprint(os.Stderr, "Error opening input file for inplace editing: %w\n", err.Error())
					// os.Remove(tempFilename);
					os.Exit(-1)
				}

				_, e := io.Copy(s.inputFile, s.outputFile)
				s.outputFile.Close()
				s.inputFile.Close()
				if e != nil {
					fmt.Fprintf(os.Stderr, "Error copying temp file back to input file: %s\nFull output is in %s", err.Error(), tempFilename)
				} else {
					os.Remove(tempFilename)
				}
			}
		}
	}
}
