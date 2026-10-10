// Convert Go's native test events to JUnit without a third-party dependency.
package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

type event struct {
	Action, Package, Test, Output string
	Elapsed                       float64
}
type failure struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}
type testCase struct {
	Name    string    `xml:"name,attr"`
	Class   string    `xml:"classname,attr"`
	Time    float64   `xml:"time,attr"`
	Failure *failure  `xml:"failure,omitempty"`
	Skipped *struct{} `xml:"skipped,omitempty"`
	Output  string    `xml:"system-out,omitempty"`
}
type suite struct {
	XMLName  xml.Name    `xml:"testsuite"`
	Name     string      `xml:"name,attr"`
	Tests    int         `xml:"tests,attr"`
	Failures int         `xml:"failures,attr"`
	Skipped  int         `xml:"skipped,attr"`
	Cases    []*testCase `xml:"testcase"`
}

func convert(input io.Reader) (suite, error) {
	result := suite{Name: "Go tests"}
	cases := map[string]*testCase{}
	packages := map[string]string{}
	failedPackages := map[string]bool{}
	decoder := json.NewDecoder(input)
	for {
		var e event
		err := decoder.Decode(&e)
		if err == io.EOF {
			break
		}
		if err != nil {
			return result, err
		}
		if e.Package == "" {
			continue
		}
		if e.Test == "" {
			packages[e.Package] += e.Output
			if e.Action == "fail" {
				failedPackages[e.Package] = true
			}
			continue
		}
		key := e.Package + "/" + e.Test
		c := cases[key]
		if c == nil {
			c = &testCase{Name: e.Test, Class: e.Package}
			cases[key] = c
		}
		c.Output += e.Output
		switch e.Action {
		case "pass":
			c.Time = e.Elapsed
		case "fail":
			c.Time = e.Elapsed
			c.Failure = &failure{Message: "Go test failed", Text: c.Output}
		case "skip":
			c.Skipped = &struct{}{}
		}
	}
	// Represent package compilation/setup failures even when no individual test ran.
	for pkg := range failedPackages {
		hasFailure := false
		for _, c := range cases {
			if c.Class == pkg && c.Failure != nil {
				hasFailure = true
			}
		}
		if !hasFailure {
			cases[pkg+"/package"] = &testCase{Name: "package build/setup", Class: pkg, Failure: &failure{Message: "Go package failed", Text: packages[pkg]}}
		}
	}
	keys := make([]string, 0, len(cases))
	for key := range cases {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		c := cases[key]
		result.Cases = append(result.Cases, c)
		result.Tests++
		if c.Failure != nil {
			result.Failures++
		}
		if c.Skipped != nil {
			result.Skipped++
		}
	}
	if result.Tests == 0 {
		return result, fmt.Errorf("no Go tests were reported")
	}
	return result, nil
}
func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: go-junit INPUT.json OUTPUT.xml")
		os.Exit(2)
	}
	input, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer input.Close()
	result, err := convert(input)
	if err != nil {
		panic(err)
	}
	output, err := xml.MarshalIndent(result, "", "  ")
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(os.Args[2], []byte(xml.Header+strings.TrimSpace(string(output))+"\n"), 0644); err != nil {
		panic(err)
	}
}
