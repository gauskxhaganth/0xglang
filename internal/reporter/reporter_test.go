package reporter

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"zerouge/internal/scanner"
	"zerouge/internal/token"
)

func TestFprettyPrintError_ScannerErrorList(t *testing.T) {
	// Buat file dummy
	tmpFile, err := os.CreateTemp("", "test_err_*.0xg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "let a = 1\nlet a::b = 2\n"
	tmpFile.WriteString(content)
	tmpFile.Close()

	var errList scanner.ErrorList
	errList.Add(token.Position{
		Filename: tmpFile.Name(),
		Line:     2,
		Column:   6,
	}, "FATAL: '::' (double colon) is forbidden on 'struct'. Use '.' for struct field and method access.")

	var buf bytes.Buffer
	FprettyPrintError(&buf, errList)

	out := buf.String()

	// Assertions
	if !strings.Contains(out, "FATAL Error") {
		t.Errorf("Expected 'FATAL Error' in output, got:\n%s", out)
	}
	// Strip ANSI for basic checks
	cleanOut := stripANSI(out)
	if !strings.Contains(cleanOut, "a::b = ") {
		t.Errorf("Expected source line in output, got:\n%s", cleanOut)
	}
	if !strings.Contains(cleanOut, "Root Cause") {
		t.Errorf("Expected 'Root Cause' in output, got:\n%s", cleanOut)
	}
	if !strings.Contains(cleanOut, "Solution") {
		t.Errorf("Expected 'Solution' in output, got:\n%s", cleanOut)
	}
}

func TestFprettyPrintError_CodegenError(t *testing.T) {
	// Buat file dummy
	tmpFile, err := os.CreateTemp("", "test_codegen_*.0xg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	content := "let a = 1\nlet a::b = 2\n"
	tmpFile.WriteString(content)
	tmpFile.Close()

	// Simulate codegen errors
	posStr := fmt.Sprintf("%s:2:6", tmpFile.Name())
	codegenErr := fmt.Errorf("%s: FATAL: '::' (double colon) is forbidden on 'struct'. Use '.' for struct field and method access.", posStr)

	var buf bytes.Buffer
	FprettyPrintError(&buf, codegenErr)

	out := buf.String()

	if !strings.Contains(out, "FATAL Error") {
		t.Errorf("Expected 'FATAL Error' in output, got:\n%s", out)
	}
	cleanOut := stripANSI(out)
	if !strings.Contains(cleanOut, "a::b = ") {
		t.Errorf("Expected source line in output, got:\n%s", cleanOut)
	}
}

func TestHighlightSyntax(t *testing.T) {
	// Red phase: test that operators and identifiers are properly colored.
	// Currently, 'highlightSyntax' only colors keywords, types, and numbers.
	input := "let x := 100 // comment"
	output := highlightSyntax(input)
	
	// 'let' should be Purple
	if !strings.Contains(output, Purple+"let"+Reset) {
		t.Errorf("Expected 'let' to be purple")
	}
	
	// '100' should be Cyan
	if !strings.Contains(output, Cyan+"100"+Reset) {
		t.Errorf("Expected '100' to be cyan")
	}
	
	// ':=' operator should be colored (each char colored separately or together)
	// Currently it emits Yellow for ':' and Yellow for '='
	if !strings.Contains(output, Yellow+":"+Reset+Yellow+"="+Reset) {
		t.Errorf("Expected ':=' to be yellow. Got: %s", output)
	}
	
	// '// comment' should be Gray
	if !strings.Contains(output, Gray+"// comment"+Reset) {
		t.Errorf("Expected comment to be gray. Got: %s", output)
	}
}

func TestTranslateGoError_GeneratedRules(t *testing.T) {
	tests := []struct {
		input       string
		expectedErr string
	}{
		{
			input:       "missing type in composite literal",
			expectedErr: "Missing type in declaration.",
		},
		{
			input:       "cannot use hello as int value in assignment",
			expectedErr: "Cannot use hello as int value in assignment.",
		},
		{
			input:       "name Foo not exported by package bar",
			expectedErr: "Name Foo not exported by cabinet bar.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			rootCause, _, matched := TranslateGoError(tc.input)
			if !matched {
				t.Fatalf("Expected '%s' to match a rule in dictionary.go, but it returned false", tc.input)
			}
			if rootCause != tc.expectedErr {
				t.Errorf("Expected root cause: %q\nGot: %q", tc.expectedErr, rootCause)
			}
		})
	}
}

func TestTranslateGoError_Solutions(t *testing.T) {
	tests := []struct {
		input            string
		expectedSolution string
	}{
		{
			input:            "name foo not exported by package bar",
			expectedSolution: "Capitalize the first letter of 'foo' to export it from the cabinet.",
		},
		{
			input:            "implicit assignment to unexported field name in struct literal of type Person",
			expectedSolution: "Capitalize the first letter of the field 'name' in the struct definition to export it.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			_, solution, matched := TranslateGoError(tc.input)
			if !matched {
				t.Fatalf("Expected '%s' to match a rule", tc.input)
			}
			if solution != tc.expectedSolution {
				t.Errorf("Expected solution: %q\nGot: %q", tc.expectedSolution, solution)
			}
		})
	}
}

func stripANSI(s string) string {
	importRegexp := `\x1b\[[0-9;]*m`
	importRe := regexp.MustCompile(importRegexp)
	return importRe.ReplaceAllString(s, "")
}
