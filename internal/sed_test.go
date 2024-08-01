// sed_test.go
// sed
//
// Original code: Copyright (c) 2009 Geoffrey Clements (MIT License)
// Modified code: Copyright (c) 2024 xplshn (3BSD License)
// For details, see the [LICENSE](https://github.com/xplshn/gosed) file at the root directory of this project
package sed

import (
	"testing"
)

func TestNewCmd(t *testing.T) {
	pieces := []byte{'4', 'x', '5', 'o', '/', '0', '/', 'g'}
	c, err := NewCmd(nil, pieces)
	if c != nil {
		t.Error("1: Got a command when we shouldn't have " + c.String())
	}
	if err == nil {
		t.Error("Didn't get an error we expected")
	} else {
		checkString(t, "Expected Unknown script command", "Unknown script command", err.Error())
	}

	// s
	pieces = []byte{'s', '/', 'o', '/', '0', '/', 'g'}
	c, err = NewCmd(nil, pieces)
	sc := c.(*SCmd)
	if sc == nil {
		t.Error("Didn't get a command that we expected")
	} else if sc.regex != "o" && len(sc.replace) == 1 && sc.replace[0] == '0' && sc.nthOccurance == -1 {
		t.Error("We didn't get the s command we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}
}

func TestNewDCmd(t *testing.T) {
	pieces := []byte{'d', '/', 'o', '/', '0', '/', 'g'}
	c, err := NewCmd(nil, pieces)
	dc := c.(*DCmd)
	if dc != nil {
		t.Error("2: Got a command when we shouldn't have " + c.String())
	}
	if err == nil {
		t.Error("Didn't get an error we expected")
	} else {
		checkString(t, "Expected: Wrong number of parameters for command", "Wrong number of parameters for command", err.Error())
	}

	pieces = []byte{'d', '/', 'd'}
	c, err = NewCmd(nil, pieces)
	dc = c.(*DCmd)
	if dc != nil {
		t.Error("3: Got a command when we shouldn't have " + c.String())
	}
	if err == nil {
		t.Error("Didn't get an error we expected")
	} else {
		checkString(t, "Expected: Wrong number of parameters for command", "Wrong number of parameters for command", err.Error())
	}

	pieces = []byte{'d'}
	c, err = NewCmd(nil, pieces)
	dc = c.(*DCmd)
	if dc == nil {
		t.Error("Didn't get a d command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}

	pieces = []byte{'$', 'd'}
	c, err = NewCmd(nil, pieces)
	dc = c.(*DCmd)
	if dc == nil {
		t.Error("Didn't get a d command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}

	pieces = []byte{'4', '5', '7', 'd'}
	c, err = NewCmd(nil, pieces)
	dc = c.(*DCmd)
	if dc == nil {
		t.Error("Didn't get a d command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}
}

func TestNewNCmd(t *testing.T) {
	pieces := []byte{'n', '/', 'o', '/', '0', '/', 'g'}
	c, err := NewCmd(nil, pieces)
	nc := c.(*NCmd)
	if nc != nil {
		t.Error("4: Got a command when we shouldn't have " + c.String())
	}
	if err == nil {
		t.Error("Didn't get an error we expected")
	}
	if err == nil {
		t.Error("Didn't get an error we expected")
	} else {
		checkString(t, "Expected: Wrong number of parameters for command", "Wrong number of parameters for command", err.Error())
	}

	pieces = []byte{'n', '/', 'd'}
	c, err = NewCmd(nil, pieces)
	nc = c.(*NCmd)
	if nc != nil {
		t.Error("5: Got a command when we shouldn't have " + c.String())
	}
	if err == nil {
		t.Error("Didn't get an error we expected")
	} else {
		checkString(t, "Expected: Wrong number of parameters for command", "Wrong number of parameters for command", err.Error())
	}

	pieces = []byte{'n'}
	c, err = NewCmd(nil, pieces)
	nc = c.(*NCmd)
	if nc == nil {
		t.Error("Didn't get a n command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}

	pieces = []byte{'$', 'n'}
	c, err = NewCmd(nil, pieces)
	nc = c.(*NCmd)
	if nc == nil {
		t.Error("Didn't get a d command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}

	pieces = []byte{'4', '5', '7', 'n'}
	c, err = NewCmd(nil, pieces)
	nc = c.(*NCmd)
	if nc == nil {
		t.Error("Didn't get a n command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}
}

func TestNewPCmd(t *testing.T) {
	pieces := []byte{'P', '/', 'o', '/', '0', '/', 'g'}
	c, err := NewCmd(nil, pieces)
	pc := c.(*PCmd)
	if pc != nil {
		t.Error("6: Got a command when we shouldn't have " + c.String())
	}
	if err == nil {
		t.Error("Didn't get an error we expected")
	}
	if err == nil {
		t.Error("Didn't get an error we expected")
	} else {
		checkString(t, "Expected: Wrong number of parameters for command", "Wrong number of parameters for command", err.Error())
	}

	pieces = []byte{'P', '/', 'd'}
	c, err = NewCmd(nil, pieces)
	pc = c.(*PCmd)
	if pc != nil {
		t.Error("7: Got a command when we shouldn't have " + c.String())
	}
	if err == nil {
		t.Error("Didn't get an error we expected")
	} else {
		checkString(t, "Expected: Wrong number of parameters for command", "Wrong number of parameters for command", err.Error())
	}

	pieces = []byte{'P'}
	c, err = NewCmd(nil, pieces)
	pc = c.(*PCmd)
	if pc == nil {
		t.Error("Didn't get a p command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}

	pieces = []byte{'$', 'P'}
	c, err = NewCmd(nil, pieces)
	pc = c.(*PCmd)
	if pc == nil {
		t.Error("Didn't get a p command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}

	pieces = []byte{'4', '5', '7', 'P'}
	c, err = NewCmd(nil, pieces)
	pc = c.(*PCmd)
	if pc == nil {
		t.Error("Didn't get a p command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}
}

func TestNewQCmd(t *testing.T) {
	pieces := []byte{'q', '/', 'o', '/', '0', '/', 'g'}
	c, err := NewCmd(nil, pieces)
	qc := c.(*QCmd)
	if err == nil {
		t.Error("Didn't get an error we expected")
	} else {
		checkString(t, "Expected: Wrong number of parameters for command", "Wrong number of parameters for command", err.Error())
	}

	pieces = []byte{'q', '/', 'q'}
	c, err = NewCmd(nil, pieces)
	qc = c.(*QCmd)
	if qc != nil {
		t.Error("9: Got a command when we shouldn't have " + c.String())
	}
	if err == nil {
		t.Error("Didn't get an error we expected")
	} else {
		checkString(t, "Expected: strconv.Atoi: parsing \"q\": invalid syntax", "strconv.Atoi: parsing \"q\": invalid syntax", err.Error())
	}

	pieces = []byte{'q'}
	c, err = NewCmd(nil, pieces)
	qc = c.(*QCmd)
	if qc == nil {
		t.Error("Didn't get a q command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}

	pieces = []byte{'q', '/', '1'}
	c, err = NewCmd(nil, pieces)
	qc = c.(*QCmd)
	if qc == nil {
		t.Error("Didn't get a q command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}

	pieces = []byte{'$', 'q'}
	c, err = NewCmd(nil, pieces)
	qc = c.(*QCmd)
	if qc == nil {
		t.Error("Didn't get a q command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}

	pieces = []byte{'4', '5', '7', 'q'}
	c, err = NewCmd(nil, pieces)
	qc = c.(*QCmd)
	if qc == nil {
		t.Error("Didn't get a d command that we expected")
	} else if err != nil {
		t.Error("Got an error we didn't expect: " + err.Error())
	}
}

func TestProcessLine(t *testing.T) {
	_s := new(Sed)
	_s.Init()
	pieces := []byte{'s', '/', 'o', '/', '0', '/', 'g'}
	c, _ := NewCmd(nil, pieces)
	_s.patternSpace = []byte{'g', 'o', 'o', 'd'}
	stop, err := c.(Cmd).processLine(_s)
	if stop {
		t.Error("Got stop when we shouldn't have")
	}
	if err != nil {
		t.Errorf("Got and error when we shouldn't have %v", err)
	}
	checkString(t, "bad global s command", "g00d", string(_s.patternSpace))

	pieces = []byte{'s', '/', 'o', '/', '0', '/', '1'}
	c, _ = NewCmd(nil, pieces)
	_s.patternSpace = []byte{'g', 'o', 'o', 'd'}
	stop, err = c.(Cmd).processLine(_s)
	if stop {
		t.Error("Got stop when we shouldn't have")
	}
	if err != nil {
		t.Errorf("Got and error when we shouldn't have %v", err)
	}
	checkString(t, "bad global s command", "g0od", string(_s.patternSpace))
}

func checkInt(t *testing.T, val, expected int, actual string) {
	if expected != val {
		t.Errorf("%d: '%d' != '%s'", val, expected, actual)
	}
}

func checkString(t *testing.T, message, expected, actual string) {
	if expected != actual {
		t.Errorf("%s: '%s' != '%s'", message, expected, actual)
	}
}
