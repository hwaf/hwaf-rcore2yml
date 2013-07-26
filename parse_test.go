package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hwaf/hwaf/hlib"
)

func TestParseLine(t *testing.T) {

	for _, v := range []struct {
		fname    string
		expected []string
	}{
		{
			fname: "testdata/package_name.txt",
			expected: []string{
				"PACKAGE", "mypackage",
			},
		},
		{
			fname: "testdata/package_name_with_tabs.txt",
			expected: []string{
				"PACKAGE", "mypackage",
			},
		},
		{
			fname: "testdata/package_cxxflags.txt",
			expected: []string{
				"PACKAGE_CXXFLAGS", "-g", "-DFOO=1", "-m64",
			},
		},
		{
			fname: "testdata/package_cxxflags_with_tabs.txt",
			expected: []string{
				"PACKAGE_CXXFLAGS", "-g", "-DFOO=1", "-m64",
			},
		},
	} {
		p, err := NewParser(v.fname)
		if err != nil {
			t.Fatalf(err.Error())
		}

		// dummy target...
		p.req.Wscript.Build.Targets = append(
			p.req.Wscript.Build.Targets,
			hlib.Target_t{},
		)

		err = p.run()
		out := p.tokens
		if !reflect.DeepEqual(out, v.expected) {
			s_expected := make([]string, 0, len(v.expected))
			for _, vv := range v.expected {
				s_expected = append(s_expected, fmt.Sprintf("%q", vv))
			}
			s_out := make([]string, 0, len(out))
			for _, vv := range out {
				s_out = append(s_out, fmt.Sprintf("%q", vv))
			}
			t.Fatalf(
				"\nexpected: %v\ngot:      %v\n",
				strings.Join(s_expected, ", "),
				strings.Join(s_out, ", "),
			)
		}
	}
}
