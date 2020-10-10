package main

import (
	"os"
	"reflect"
	"testing"
)

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
	}

	for _, tCase := range testMap {
		args := Args{IpV4Mask: tCase.V4Mask, IpV6Mask: tCase.V6Mask, Columns: []uint{0}}
		maskedLine := handleLine(tCase.Input, args)
		if maskedLine != tCase.Expected {
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
		args := Args{Increment: tCase.Increment, IpV4Mask: 12, IpV6Mask: 84, Columns: []uint{0}}
		maskedLine := handleLine(tCase.Input, args)
		if maskedLine != tCase.Expected {
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
		args := Args{Columns: tCase.Columns, IpV4Mask: 12, IpV6Mask: 84}
		maskedLine := handleLine(tCase.Input, args)
		if maskedLine != tCase.Expected {
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
