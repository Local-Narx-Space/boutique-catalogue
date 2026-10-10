package main

import (
	"strings"
	"testing"
)

func TestReportsFailureAndSkip(t *testing.T) {
	events := `{"Action":"run","Package":"example","Test":"bad"}
{"Action":"output","Package":"example","Test":"bad","Output":"assertion failed"}
{"Action":"fail","Package":"example","Test":"bad","Elapsed":0.1}
{"Action":"skip","Package":"example","Test":"optional"}`
	report, err := convert(strings.NewReader(events))
	if err != nil || report.Tests != 2 || report.Failures != 1 || report.Skipped != 1 {
		t.Fatalf("lost test outcomes: %#v, %v", report, err)
	}
	if report.Cases[0].Failure.Text != "assertion failed" {
		t.Fatal("failure diagnostic missing")
	}
}
func TestReportsPackageBuildFailure(t *testing.T) {
	report, err := convert(strings.NewReader(`{"Action":"output","Package":"example","Output":"compile error"}
{"Action":"fail","Package":"example"}`))
	if err != nil || report.Tests != 1 || report.Failures != 1 {
		t.Fatalf("package failure lost: %#v, %v", report, err)
	}
}
func TestRejectsEmptyOrInvalidInput(t *testing.T) {
	for _, input := range []string{"", "not json", `{"Action":"pass","Package":"example"}`} {
		if _, err := convert(strings.NewReader(input)); err == nil {
			t.Fatalf("accepted invalid/empty test run: %q", input)
		}
	}
}
