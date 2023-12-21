package pbn

import (
	"errors"
	"github.com/alecthomas/participle/v2"
	"reflect"
	"testing"
)

var (
	pbn = `% Export
% PBN 2.1
[Event "Game 1"]
[Site "Site 1"]

[Event "Game with tables"]
[TestTable "x;y;z"]
 x - z
 1 2 "Test"

[Event "Game with tags that occurs more than once"]
[Test "Test 1"]
[Test "Test 2"]
`
)

func TestPbnParser(t *testing.T) {
	p, err := parser.ParseString("", pbn)
	if err != nil {
		t.Errorf("error parsing pbn: %v", err)
	}
	if len(p.Game) != 3 {
		t.Errorf("Expected %v games in pbn, got %v", 3, len(p.Game))
	}

	expectedTableData := TableData{{"x", "-", "z"}, {"1", "2", "Test"}}
	data := p.Game[1].Tables["TestTable"].Data
	if reflect.DeepEqual(data, expectedTableData) == false {
		t.Errorf("Expected %+v, got %+v", expectedTableData, expectedTableData)
	}

	if p.Game[2].Tags["Test"].Value != "Test 1" {
		t.Errorf("Tag occuring more than one should ignore the second value")
	}
}

func TestEscapedLineParser(t *testing.T) {
	tests := map[string]struct {
		input string
	}{
		"empty %":     {"%"},
		"single line": {"% Export"},
		"two lines":   {"% Line 1\r\n% Line 2"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := escapedParser.ParseString("", tc.input, participle.AllowTrailing(true))
			if err != nil {
				t.Fatalf("Could not parse tag: %v", err)
			}
		})
	}
}

func TestTagParser(t *testing.T) {
	tests := map[string]struct {
		input     string
		expected  Tag
		shouldErr bool
	}{
		"single tag":           {`[Name "TagValue"]`, Tag{"Name", "TagValue"}, false},
		"strange letters":      {`[Name "Välüe"]`, Tag{"Name", "Välüe"}, false},
		"invalid name token":   {`[Name! "TagValue"]`, Tag{}, true},
		"value not a string":   {`[Name TagValue]`, Tag{}, true},
		"missing name":         {`["TagValue"]`, Tag{}, true},
		"missing value":        {`[Name]`, Tag{}, true},
		"name not capitalized": {`[name "TagValue"]`, Tag{}, true},
		"empty tag":            {`[]`, Tag{}, true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			actual, err := tagParser.ParseString("", tc.input)
			if tc.shouldErr {
				var e *participle.UnexpectedTokenError
				if !errors.As(err, &e) {
					t.Fatalf("Expected a unexpected token error, got %+v", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Could not parse tag: %v", err)
			}
			if !actual.Equal(tc.expected) {
				t.Errorf("Expected %v, got %+v", tc.expected, actual)
			}
		})
	}
}
