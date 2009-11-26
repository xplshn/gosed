package sed

import (
	"os";
	"fmt";
	"strconv";
	"regexp";
)

const debug = false

type cmd struct {
	operation	string;
	parameter	string;
	replace		string;
	flag		string;
	// used for the s command
	re	*regexp.Regexp;
}

func (c *cmd) String() string {
	return fmt.Sprintf("cmd {operation: %s parameter: %s replace: %s flag: %s}", c.operation, c.parameter, c.replace, c.flag)
}

func NewCmd(pieces []string) (c *cmd, err os.Error) {
	err = nil;
	c = new(cmd);
	c.operation = pieces[0];
	switch c.operation {
	case "s":
		if len(pieces) != 4 {
			return nil, os.ErrorString("invalid script line")
		}
		c.parameter = pieces[1];
		c.replace = pieces[2];
		c.flag = pieces[3];
		if len(c.parameter) == 0 {
			return nil, os.ErrorString("Regular expression in s command can't be zero length.")
		}
		c.re, err = regexp.Compile(c.parameter);
		if err != nil {
			c = nil
		}
	case "q":
		if len(pieces) != 2 && len(pieces) != 1 {
			return nil, os.ErrorString("invalid script line")
		}
		if len(pieces) == 2 {
			c.parameter = pieces[1]
		}
	case "d":
		// do nothing else
	case "P":
		// do nothing else
	case "n":
		// do nothing else
	default:
		c, err = nil, os.ErrorString("unknown script command")
	}
	if debug {
		fmt.Println("Created: ", c)
	}
	return c, err;
}

func (c *cmd) processLine(line string) (processSpace string, stop bool, err os.Error) {
	// setup defailt return values
	processSpace, stop, err = line, false, nil;
	switch c.operation {
	case "s":
		if debug {
			fmt.Println("s cmd: ", c)
		}
		switch c.flag {
		case "g":
			if debug {
				fmt.Println("cmd s with global replace")
			}
			processSpace = c.re.ReplaceAllString(line, c.replace);
		default:
			// a numeric flag command
			count := 1;
			if len(c.flag) > 0 {
				newCount, err := strconv.Atoi(c.flag);
				if err != nil {
					processSpace, stop, err = "", true, os.ErrorString("Invalid flag for s command");
					return;
				}
				count = newCount;
			}
			if debug {
				fmt.Println("cmd s with ", count, " replace")
			}
			processSpace = "";
			for {
				if count <= 0 {
					processSpace += line;
					return;
				}
				lineLength := len(line);
				matches := c.re.ExecuteString(line);
				if len(matches) == 0 {
					processSpace += line;
					return;
				}
				processSpace += line[0:matches[0]] + c.replace;
				line = line[matches[1]:lineLength];
				count--;
			}
		}
	case "q":
		if debug {
			fmt.Println("q cmd: ", c)
		}
		exitCode, err := strconv.Atoi(c.parameter);
		if err == nil {
			os.Exit(exitCode)
		} else {
			os.Exit(0)
		}
	case "P":
		if debug {
			fmt.Println("p cmd: ", c)
		}
		fmt.Fprintln(outputFile, line);
	case "d":
		// delete the patternSpace and go onto next line
		stop = true;
		line = "";
	case "n":
		if !*quiet {
			printPatternSpace(line)
		}
		line = "";
		stop = true;
	default:
		line, stop, err = "", true, os.ErrorString("unknown script command")
	}
	if debug {
		fmt.Println("processLine returns: ", processSpace)
	}
	return processSpace, stop, nil;
}