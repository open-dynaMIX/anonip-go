package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"regexp"
	"testing"
	"testing/iotest"
)

func TestMain(m *testing.M) {
	rc := m.Run()

	// rc 0 means we've passed,
	// and CoverMode will be non empty if run with -cover
	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < 1.0 { // enforce 100% coverage
			fmt.Println("Tests passed but coverage failed at", c)
			rc = -1
		}
	}
	os.Exit(rc)
}

func GetDefaultArgs() Args {
	oldOsArgs := os.Args
	defer func() { os.Args = oldOsArgs }()
	os.Args = []string{"anonip"}
	args, _, _ := parseArgs()
	return args
}

func TestHandleLine(t *testing.T) {
	type TestCase struct {
		Input    string
		Expected string
		V4Mask   int
		V6Mask   int
	}
	testMap := []TestCase{
		{
			Input:    "3.3.3.3",
			Expected: "3.3.0.0",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			Expected: "2001:db8:85a0::",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "192.168.100.200",
			Expected: "192.168.96.0",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "192.168.100.200:80",
			Expected: "192.168.96.0:80",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "192.168.100.200]",
			Expected: "192.168.96.0]",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "192.168.100.200:80]",
			Expected: "192.168.96.0:80]",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "192.168.100.200",
			Expected: "192.168.100.200",
			V4Mask:   0,
			V6Mask:   84,
		},
		{
			Input:    "192.168.100.200",
			Expected: "192.168.100.192",
			V4Mask:   4,
			V6Mask:   84,
		},
		{
			Input:    "192.168.100.200",
			Expected: "192.168.100.0",
			V4Mask:   8,
			V6Mask:   84,
		},
		{
			Input:    "192.168.100.200",
			Expected: "192.0.0.0",
			V4Mask:   24,
			V6Mask:   84,
		},
		{
			Input:    "192.168.100.200",
			Expected: "0.0.0.0",
			V4Mask:   32,
			V6Mask:   84,
		},
		{
			Input:    "no_ip_address",
			Expected: "no_ip_address",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			Expected: "2001:db8:85a0::",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:443",
			Expected: "[2001:db8:85a0::]:443",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]",
			Expected: "[2001:db8:85a0::]",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]]",
			Expected: "[2001:db8:85a0::]]",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:443]",
			Expected: "[2001:db8:85a0::]:443]",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			Expected: "2001:db8:85a3::8a2e:370:7334",
			V4Mask:   12,
			V6Mask:   0,
		},
		{
			Input:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			Expected: "2001:db8:85a3::8a2e:370:7330",
			V4Mask:   12,
			V6Mask:   4,
		},
		{
			Input:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			Expected: "2001:db8:85a3::8a2e:370:7300",
			V4Mask:   12,
			V6Mask:   8,
		},
		{
			Input:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			Expected: "2001:db8:85a3::8a2e:300:0",
			V4Mask:   12,
			V6Mask:   24,
		},
		{
			Input:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			Expected: "2001:db8:85a3::8a2e:0:0",
			V4Mask:   12,
			V6Mask:   32,
		},
		{
			Input:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			Expected: "2001:db8:85a3::",
			V4Mask:   12,
			V6Mask:   62,
		},
		{
			Input:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			Expected: "::",
			V4Mask:   12,
			V6Mask:   128,
		},
		{
			Input:    "   foo",
			Expected: "   foo",
			V4Mask:   12,
			V6Mask:   84,
		},
		{
			Input:    "",
			Expected: "",
			V4Mask:   12,
			V6Mask:   84,
		},
	}

	for _, tCase := range testMap {
		channel := make(chan string)
		args := GetDefaultArgs()
		args.IpV4Mask, args.IpV6Mask = tCase.V4Mask, tCase.V6Mask
		go handleLine(tCase.Input, args, channel)
		if maskedLine := <-channel; maskedLine != tCase.Expected {
			t.Errorf("Failing input: %+v\nReceived output: %v", tCase, maskedLine)
		}
	}
}

func TestIncrement(t *testing.T) {
	type TestCase struct {
		Input     string
		Increment uint
		Expected  string
	}
	testMap := []TestCase{
		{
			Input:     "192.168.100.200",
			Increment: 3,
			Expected:  "192.168.96.3",
		},
		{
			Input:     "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			Increment: 7,
			Expected:  "2001:db8:85a0::7",
		},
	}
	for _, tCase := range testMap {
		channel := make(chan string)
		args := GetDefaultArgs()
		args.Increment = tCase.Increment
		go handleLine(tCase.Input, args, channel)
		if maskedLine := <-channel; maskedLine != tCase.Expected {
			t.Errorf("Failing input: %+v\nReceived output: %v", tCase, maskedLine)
		}
	}
}

func TestColumns(t *testing.T) {
	type TestCase struct {
		Input    string
		Columns  []uint
		Expected string
	}
	testMap := []TestCase{
		{
			Input:    "192.168.100.200 some string with öéäü",
			Columns:  []uint{0},
			Expected: "192.168.96.0 some string with öéäü",
		},
		{
			Input:    "some 192.168.100.200 string with öéäü",
			Columns:  []uint{1},
			Expected: "some 192.168.96.0 string with öéäü",
		},
		{
			Input:    "some string 192.168.100.200 with öéäü",
			Columns:  []uint{2},
			Expected: "some string 192.168.96.0 with öéäü",
		},
		{
			Input:    "192.168.100.200 192.168.11.222 192.168.123.234",
			Columns:  []uint{0, 1, 2},
			Expected: "192.168.96.0 192.168.0.0 192.168.112.0",
		},
		{
			Input:    "192.168.100.200 192.168.11.222 192.168.123.234",
			Columns:  []uint{9999},
			Expected: "192.168.100.200 192.168.11.222 192.168.123.234",
		},
	}
	for _, tCase := range testMap {
		channel := make(chan string)
		args := GetDefaultArgs()
		args.Columns = tCase.Columns
		go handleLine(tCase.Input, args, channel)
		if maskedLine := <-channel; maskedLine != tCase.Expected {
			t.Errorf("Failing input: %+v\nReceived output: %v", tCase, maskedLine)
		}
	}
}

func TestArgsColumns(t *testing.T) {
	type TestCase struct {
		Input    []string
		Expected []uint
		Success  bool
	}
	testMap := []TestCase{
		{
			Input:    []string{""},
			Expected: []uint{0},
			Success:  true,
		},
		{
			Input:    []string{"-c", "5"},
			Expected: []uint{4},
			Success:  true,
		},
		{
			Input:    []string{"-c", "2", "5"},
			Expected: []uint{1, 4},
			Success:  true,
		},
		{
			Input:    []string{"-c", "0"},
			Expected: []uint{0},
			Success:  false,
		},
	}

	defer func() { os.Args = []string{"anonip"} }()

	for _, tCase := range testMap {
		os.Args = []string{"anonip"}
		if tCase.Input[0] != "" {
			os.Args = append(os.Args, tCase.Input...)
		}
		args, _, err := parseArgs()
		if err != nil && tCase.Success {
			t.Errorf("Failed with input: %v", tCase.Input)
		}
		if !reflect.DeepEqual(args.Columns, tCase.Expected) {
			t.Errorf("Test failed")
		}
	}
}

func TestArgsIPMasks(t *testing.T) {
	type TestCase struct {
		Input   []string
		Success bool
	}
	testMap := []TestCase{
		{
			Input:   []string{"-4", "12", "-6", "84"},
			Success: true,
		},
		{
			Input:   []string{"-4", "-1", "-6", "130"},
			Success: false,
		},
		{
			Input:   []string{"-4", "33", "-6", "84"},
			Success: false,
		},
		{
			Input:   []string{"-4", "12", "-6", "129"},
			Success: false,
		},
	}

	defer func() { os.Args = []string{"anonip"} }()

	for _, tCase := range testMap {
		os.Args = []string{"anonip"}
		if tCase.Input[0] != "" {
			os.Args = append(os.Args, tCase.Input...)
		}
		_, _, err := parseArgs()
		if err == nil && !tCase.Success {
			t.Errorf("Should have failed with input: %v", tCase.Input)
		} else if err != nil && tCase.Success {
			t.Errorf("Should not have failed with input: %v", tCase.Input)
		}
	}
}

func _TestMain(Input []byte, Expected string, Regex string, t *testing.T) {
	// create a copy of the old stdin and stdout
	oldStdin := os.Stdin
	oldStdout := os.Stdout

	// Create pipes for monkey patching stdin and stdout
	stdoutPipeRead, stdoutPipeWrite, _ := os.Pipe()
	stdinPipeRead, stdinPipeWrite, _ := os.Pipe()

	// reassign stdin and stdout
	logReader = stdinPipeRead
	logWriter = stdoutPipeWrite

	// make sure to clean up afterwards
	defer func() {
		os.Args = []string{"anonip"}
		logReader = oldStdin
		logWriter = oldStdout
	}()

	os.Args = []string{"anonip"}
	if Regex != "" {
		os.Args = append(os.Args, []string{"--regex", Regex}...)
	}
	// Write input to stdin pipe
	if _, err := stdinPipeWrite.Write(Input); err != nil {
		log.Fatal(err)
	}

	go main()

	// read the output from the stdout pipe
	buf := make([]byte, 1024)
	n, err := stdoutPipeRead.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	output := string(buf[:n])

	if output != Expected {
		t.Errorf("Wanted: %v, Got: %v", Expected, output)
	}
}

func TestMainSuccess(t *testing.T) {
	type TestCase struct {
		Input    []byte
		Expected string
		Regex    string
	}
	testMap := []TestCase{
		{
			Input:    []byte("192.168.100.200\n"),
			Expected: "192.168.96.0\n",
		},
		{
			Input:    []byte("2001:0db8:85a3:0000:0000:8a2e:0370:7334\n"),
			Expected: "2001:db8:85a0::\n",
		},
		{
			Input:    []byte("bla 192.168.100.200\n"),
			Expected: "bla 192.168.96.0\n",
			Regex:    "^bla (.*)",
		},
	}

	for _, tCase := range testMap {
		_TestMain(tCase.Input, tCase.Expected, tCase.Regex, t)
	}
}

func TestMainFail(t *testing.T) {
	testMap := [][]string{
		{"-c", "0"},
		{"-4", "33"},
		{"-6", "-1"},
		{"-o"},
		{"--input"},
		{"--regex", "\\8"},
	}

	// ignore stderr in order to keep the log clean
	oldStderr := os.Stderr
	os.Stderr, _ = os.Open("/dev/null")

	tempDir, err := ioutil.TempDir("", "tempLog")
	if err != nil {
		log.Fatal(err)
	}

	// patched exit function
	var got int
	testOsExit := func(code int) {
		got = code
	}

	// create a copy of the old value
	oldOsExit := osExit

	// reassign osExit
	osExit = testOsExit

	// restore previous state after the test
	defer func() {
		os.Args = []string{"anonip"}
		os.Stderr = oldStderr
		osExit = oldOsExit
		err := os.Remove(tempDir)
		if err != nil {
			log.Fatal("error:", err)
		}
	}()

	for _, tCase := range testMap {
		// setup args
		if len(tCase) == 1 {
			tCase = append(tCase, tempDir)
		}
		os.Args = append([]string{"anonip"}, tCase...)

		main()

		// Check if exit code has been called
		_got := got
		got = 0
		if _got != -1 {
			t.Errorf("Expected exit code: -1, got: %d", _got)
		}
	}
}

func TestRunFail(t *testing.T) {
	// patched exit function
	var got int
	testOsExit := func(code int) {
		got = code
	}

	// create a copy of the old value
	oldOsExit := osExit

	// ignore stderr in order to keep the log clean
	oldStderr := os.Stderr
	os.Stderr, _ = os.Open("/dev/null")

	// restore previous state after the test
	defer func() {
		osExit = oldOsExit
		os.Stderr = oldStderr
	}()

	// reassign osExit
	osExit = testOsExit

	logReader = iotest.TimeoutReader(bytes.NewReader([]byte("test")))

	args := GetDefaultArgs()

	run(args)

	if got != -1 {
		t.Errorf("Expected exit code: -1, got: %d", got)
	}
}

func TestDelimiter(t *testing.T) {
	type TestCase struct {
		Input     string
		Delimiter string
		Expected  string
	}
	testMap := []TestCase{
		{
			Input:     "192.168.100.200;some;string;with;öéäü",
			Delimiter: ";",
			Expected:  "192.168.96.0;some;string;with;öéäü",
		},
		{
			Input:     "192.168.100.200 some string with öéäü",
			Delimiter: ";",
			Expected:  "192.168.100.200 some string with öéäü",
		},
	}
	for _, tCase := range testMap {
		channel := make(chan string)
		args := GetDefaultArgs()
		args.Delimiter = tCase.Delimiter
		go handleLine(tCase.Input, args, channel)
		if maskedLine := <-channel; maskedLine != tCase.Expected {
			t.Errorf("Failing input: %+v\nReceived output: %v", tCase, maskedLine)
		}
	}
}

func TestReplace(t *testing.T) {
	type TestCase struct {
		Input    string
		Replace  *string
		Expected string
	}
	replaceString := "replaceIt"
	testMap := []TestCase{
		{
			Input:    "some string without IP",
			Replace:  nil,
			Expected: "some string without IP",
		},
		{
			Input:    "some string without IP",
			Replace:  &replaceString,
			Expected: "replaceIt string without IP",
		},
	}
	for _, tCase := range testMap {
		channel := make(chan string)
		args := GetDefaultArgs()
		args.Replace = tCase.Replace
		go handleLine(tCase.Input, args, channel)
		if maskedLine := <-channel; maskedLine != tCase.Expected {
			t.Errorf("Failing input: %+v\nReceived output: %v", tCase, maskedLine)
		}
	}
}

func TestSkipPrivate(t *testing.T) {
	type TestCase struct {
		Input    string
		Expected string
	}
	testMap := []TestCase{
		{
			Input:    "10.0.0.1",
			Expected: "10.0.0.1",
		},
		{
			Input:    "3.3.3.3",
			Expected: "3.3.0.0",
		},
		{
			Input:    "169.254.0.1",
			Expected: "169.254.0.1",
		},
	}
	for _, tCase := range testMap {
		channel := make(chan string)
		args := GetDefaultArgs()
		args.SkipPrivate = true
		initPrivateIPBlocks()
		go handleLine(tCase.Input, args, channel)
		if maskedLine := <-channel; maskedLine != tCase.Expected {
			t.Errorf("Failing input: %+v\nReceived output: %v", tCase, maskedLine)
		}
	}
}

func TestFailInitPrivateIPBlocks(t *testing.T) {
	// patched exit function
	var got int
	testOsExit := func(code int) {
		got = code
	}

	// create a copy of the old value
	oldOsExit := osExit

	// ignore stderr in order to keep the log clean
	oldStderr := os.Stderr
	os.Stderr, _ = os.Open("/dev/null")

	// restore previous state after the test
	defer func() {
		osExit = oldOsExit
		os.Stderr = oldStderr
	}()

	// reassign osExit
	osExit = testOsExit

	privateIPBlocksStrings = []string{
		"no valid CIDR",
	}

	args := GetDefaultArgs()
	args.SkipPrivate = true

	run(args)

	if got != 2 {
		t.Errorf("Expected exit code: 2, got: %d", got)
	}
}

func TestRegexMatching(t *testing.T) {
	type TestCase struct {
		Input    string
		Expected string
		Regex    string
	}
	testMap := []TestCase{
		{
			Input:    "3.3.3.3 - - [20/May/2015:21:05:01 +0000] \"GET / HTTP/1.1\" 200 13358 \"-\" \"useragent\"\n",
			Expected: "3.3.0.0 - - [20/May/2015:21:05:01 +0000] \"GET / HTTP/1.1\" 200 13358 \"-\" \"useragent\"\n",
			Regex:    "(?:^(.*) - - |.* - somefixedstring: (.*) - .* - (.*))",
		},
		{
			Input:    "blabla/ 3.3.3.3 /blublu",
			Expected: "blabla/ 3.3.0.0 /blublu",
			Regex:    "^blabla/ (.*) /blublu$",
		},
	}

	for _, tCase := range testMap {
		channel := make(chan string)
		args := GetDefaultArgs()
		args.Regex = tCase.Regex
		regex = regexp.MustCompile(args.Regex)
		go handleLine(tCase.Input, args, channel)
		if maskedLine := <-channel; maskedLine != tCase.Expected {
			t.Errorf("Failing input: %+v\nReceived output: %v", tCase, maskedLine)
		}
	}
}
