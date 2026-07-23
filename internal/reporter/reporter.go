package reporter

import (
	"fmt"
	"io"
	"os"
	"strings"
	"orez/internal/scanner"
)

// ANSI Color Codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
	Bold   = "\033[1m"
)

var (
	kwMap = map[string]bool{
		"let": true, "def": true, "class": true, "struct": true, "require": true, 
		"retain": true, "cabinet": true, "if": true, "end": true, "elsif": true, 
		"else": true, "return": true, "for": true, "in": true, "case": true, 
		"when": true, "go": true, "defer": true, "select": true,
	}
	typeMap = map[string]bool{
		"String": true, "Int": true, "Int8": true, "Int16": true, "Int32": true, 
		"Int64": true, "Uint": true, "Uint8": true, "Uint16": true, "Uint32": true, 
		"Uint64": true, "Float32": true, "Float64": true, "Bool": true, "Byte": true, 
		"Rune": true, "Any": true, "Array": true, "Rigid": true, "Map": true, "Channel": true,
	}
)

func highlightSyntax(s string) string {
	var out strings.Builder
	inString := false
	inComment := false
	var word strings.Builder

	flushWord := func() {
		if word.Len() > 0 {
			w := word.String()
			if kwMap[w] {
				out.WriteString(Purple + w + Reset)
			} else if typeMap[w] {
				out.WriteString(Yellow + w + Reset)
			} else if w[0] >= '0' && w[0] <= '9' {
				out.WriteString(Cyan + w + Reset)
			} else {
				out.WriteString(w)
			}
			word.Reset()
		}
	}

	for i := 0; i < len(s); i++ {
		if inComment {
			out.WriteByte(s[i])
			continue
		}
		if inString {
			out.WriteByte(s[i])
			if s[i] == '"' {
				out.WriteString(Reset)
				inString = false
			}
			continue
		}

		if i+1 < len(s) && s[i] == '/' && s[i+1] == '/' {
			flushWord()
			out.WriteString(Gray)
			inComment = true
			out.WriteByte('/')
			out.WriteByte('/')
			i++
			continue
		}

		if s[i] == '"' {
			flushWord()
			out.WriteString(Green)
			out.WriteByte('"')
			inString = true
			continue
		}

		isAlphaNum := (s[i] >= 'a' && s[i] <= 'z') || (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= '0' && s[i] <= '9') || s[i] == '_' || s[i] == '.'
		if isAlphaNum {
			word.WriteByte(s[i])
		} else {
			flushWord()
			// Punctuation / Operator coloring
			switch s[i] {
			case ':', '=', '+', '-', '*', '/', '%', '!', '<', '>', '&', '|', '^':
				out.WriteString(Yellow)
				out.WriteByte(s[i])
				out.WriteString(Reset)
			default:
				out.WriteByte(s[i])
			}
		}
	}
	flushWord()
	if inComment {
		out.WriteString(Reset)
	}
	return out.String()
}

// PrettyPrintError formats and prints compiler errors beautifully to stdout.
func PrettyPrintError(err error) {
	FprettyPrintError(os.Stdout, err)
}

// FprettyPrintError formats and prints compiler errors beautifully to the provided io.Writer.
func FprettyPrintError(w io.Writer, err error) {
	if list, ok := err.(scanner.ErrorList); ok {
		for _, e := range list {
			PrintSingleError(w, e.Pos.Filename, e.Pos.Line, e.Pos.Column, e.Msg)
		}
		return
	}

	// Handle wrapped errors from codegen like "%s: FATAL: ..."
	errMsg := err.Error()
	if strings.Contains(errMsg, ": FATAL:") {
		parts := strings.SplitN(errMsg, ": FATAL:", 2)
		if len(parts) == 2 {
			posParts := strings.Split(parts[0], ":")
			if len(posParts) >= 3 {
				filename := strings.Join(posParts[:len(posParts)-2], ":")
				var line, col int
				fmt.Sscanf(posParts[len(posParts)-2], "%d", &line)
				fmt.Sscanf(posParts[len(posParts)-1], "%d", &col)
				PrintSingleError(w, filename, line, col, "FATAL:"+parts[1])
				return
			}
		}
	}

	// Fallback to plain print
	fmt.Fprintln(w, Red+Bold+err.Error()+Reset)
}

func PrintSingleError(w io.Writer, filename string, line, col int, fullMsg string) {
	fullMsg = strings.TrimSpace(fullMsg)
	
	rootCause := fullMsg
	solution := ""
	
	if strings.HasPrefix(rootCause, "FATAL: ") {
		rootCause = rootCause[7:] 
	}

	if idx := strings.Index(rootCause, ". Use "); idx != -1 {
		solution = "Use " + rootCause[idx+6:]
		rootCause = rootCause[:idx+1]
	} else if idx := strings.Index(rootCause, ". "); idx != -1 && len(rootCause) > idx+2 {
		solution = rootCause[idx+2:]
		rootCause = rootCause[:idx+1]
	}

	// Use Dictionary Translation if there is no solution or the format is still native Go
	translatedCause, dictSolution, found := TranslateGoError(rootCause)
	if found {
		rootCause = translatedCause
		if solution == "" {
			solution = dictSolution
		}
	}

	if solution == "" {
		// Hardcoded fallback specifically for internal 0xg parser
		if strings.Contains(rootCause, "curly braces '{' are forbidden") {
			solution = "Remove the '{' brace. 0xg uses block-keywords and 'end' without opening braces."
		} else if strings.Contains(rootCause, "Execution of 'go' must be followed") || strings.Contains(rootCause, "Execution of 'defer' must be followed") {
			solution = "Add a valid function call (e.g., 'go myFunc()') after the keyword."
		} else if strings.Contains(rootCause, "Expected package string") || strings.Contains(rootCause, "Expected 'retain'") {
			solution = "Use block format: require \\n \"pkg\" \\n retain"
		} else if strings.Contains(rootCause, "expected identifier after '@'") {
			solution = "Provide a valid instance variable name after '@' (e.g., '@name')."
		} else if strings.Contains(rootCause, "expected class name after 'class'") || strings.Contains(rootCause, "Expected struct name") {
			solution = "Provide a valid identifier name for the struct/class."
		} else {
			solution = "Review the syntax or logic and ensure it complies with 0xg rules."
		}
	}

	fmt.Fprintf(w, "\n%s%sFATAL Error%s\n", Bold, Red, Reset)
	fmt.Fprintf(w, " %s╭───%s `%s:%s%d%s:%s%d%s`\n", Blue, Reset, filename, Yellow, line, Reset, Cyan, col, Reset)
	fmt.Fprintf(w, " %s│%s\n", Blue, Reset)

	sourceCode := ""
	srcBytes, err := os.ReadFile(filename)
	if err == nil {
		sourceCode = string(srcBytes)
	}

	lines := strings.Split(sourceCode, "\n")
	
	if line > 1 && line-2 < len(lines) {
		prevOrig := lines[line-2]
		if len(prevOrig) > 80 {
			prevOrig = prevOrig[:80] + "..."
		}
		prevLine := highlightSyntax(strings.ReplaceAll(prevOrig, "\t", "    "))
		fmt.Fprintf(w, " %s│%s %4d %s│%s %s\n", Blue, Reset, line-1, Blue, Reset, prevLine)
	}

	if line > 0 && line-1 < len(lines) {
		origLine := lines[line-1]
		
		// Setup term width limit
		const maxLen = 80
		
		// If line is too long, we truncate/wrap it
		if len(origLine) > maxLen {
			// Print truncated/wrapped version to prevent breaking the border
			errLineStr := highlightSyntax(strings.ReplaceAll(origLine[:maxLen]+"...", "\t", "    "))
			fmt.Fprintf(w, " %s│%s %s%4d%s %s│%s %s\n", Blue, Reset, Bold, line, Reset, Blue, Reset, errLineStr)
			
			// Adjust col position if it's beyond maxLen
			visualCol := 0
			for i := 0; i < col-1 && i < len(origLine); i++ {
				if origLine[i] == '\t' {
					visualCol += 4
				} else {
					visualCol++
				}
			}
			
			if visualCol > maxLen {
				visualCol = maxLen + 1 // Point to the "..."
			}
			
			padding := strings.Repeat(" ", visualCol)
			fmt.Fprintf(w, " %s│%s      %s│%s %s%s^^%s\n", Blue, Reset, Blue, Reset, padding, Red, Reset)
			fmt.Fprintf(w, " %s│%s      %s│%s %s╰── Root Cause:%s\n", Blue, Reset, Blue, Reset, Red, Reset)
			fmt.Fprintf(w, " %s│%s      %s│%s     %s\n", Blue, Reset, Blue, Reset, rootCause)
		} else {
			// Normal printing for short lines
			errLineStr := highlightSyntax(strings.ReplaceAll(origLine, "\t", "    "))
			fmt.Fprintf(w, " %s│%s %s%4d%s %s│%s %s\n", Blue, Reset, Bold, line, Reset, Blue, Reset, errLineStr)
			
			padding := "    "
			if col > 0 {
				visualCol := 0
				for i := 0; i < col-1 && i < len(origLine); i++ {
					if origLine[i] == '\t' {
						visualCol += 4
					} else {
						visualCol++
					}
				}
				padding = strings.Repeat(" ", visualCol)
			}
			
			fmt.Fprintf(w, " %s│%s      %s│%s %s%s^^%s\n", Blue, Reset, Blue, Reset, padding, Red, Reset)
			fmt.Fprintf(w, " %s│%s      %s│%s %s╰── Root Cause:%s\n", Blue, Reset, Blue, Reset, Red, Reset)
			fmt.Fprintf(w, " %s│%s      %s│%s     %s\n", Blue, Reset, Blue, Reset, rootCause)
		}
	} else {
		fmt.Fprintf(w, " %s│%s      %s│%s %s╰── Root Cause:%s\n", Blue, Reset, Blue, Reset, Red, Reset)
		fmt.Fprintf(w, " %s│%s      %s│%s     %s\n", Blue, Reset, Blue, Reset, rootCause)
	}

	fmt.Fprintf(w, " %s│%s\n", Blue, Reset)
	if solution != "" {
		fmt.Fprintf(w, " %s╰───%s %sSolution:%s %s\n", Blue, Reset, Green, Reset, solution)
	} else {
		fmt.Fprintf(w, " %s╰───%s\n", Blue, Reset)
	}
	fmt.Fprintln(w)
}
