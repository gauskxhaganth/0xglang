package compiler

import (
	"bytes"
	"strings"
	"testing"
)

// Layer 7 Test: Background Executor Test (Compiler Proxy End-to-End)
func TestCompilerProxyRun(t *testing.T) {
	// Pure 0xg script
	src := `cabinet main

require "fmt"

def main()
	let result = 15 + 5
	fmt.Println(result)
end`

	var out bytes.Buffer
	// Run script and capture Standard Output
	err := RunSource([]byte(src), &out)
	if err != nil {
		t.Fatalf("RunSource error: %v, out: %s", err, out.String())
	}
}

// Layer 8 Test: Source Mapping (Compiler Directive //line)
func TestSourceMappingError(t *testing.T) {
	src := `cabinet main
def main()
	let x = "string" + 5
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err == nil {
		t.Fatalf("RED PHASE: Expected compilation to fail (Type Error), but it succeeded")
	}

	errStr := out.String()
	if !strings.Contains(errStr, "main.0xg:3") {
		t.Errorf("RED PHASE: Compiler error message does not point to main.0xg:3. Output error: %s", errStr)
	}
}

// Layer 9 Test: Function Call (CallExpr)
func TestCallExpr(t *testing.T) {
	src := `cabinet main
def main()
	let a = 10
	let b = 20
	println(a + b)
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("RED PHASE: Failed to execute CallExpr. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "30") {
		t.Errorf("RED PHASE: Expected to print '30', got: '%s'", errStr)
	}
}

// Layer 10 Test: Library Interoperability (RequireDecl & SelectorExpr)
func TestRequireStmt(t *testing.T) {
	src := `cabinet main

require "fmt"

def main()
	fmt.Println("Hello from 0xg!")
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("RED PHASE: Failed to execute RequireDecl. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "Hello from 0xg!") {
		t.Errorf("RED PHASE: Expected to print 'Hello from 0xg!', got: '%s'", errStr)
	}
}

// Layer 11 Test: Value Return (ReturnStmt)
func TestReturnStmt(t *testing.T) {
	src := `cabinet main
require "fmt"

def Calculate(a Int, b Int) Int
	return a * b
end

def main()
	let result = Calculate(5, 5)
	fmt.Println(result)
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("RED PHASE: Failed to execute ReturnStmt. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "25") {
		t.Errorf("RED PHASE: Expected to print '25', got: '%s'", errStr)
	}
}

// Layer 12 Test: Object Declaration (Struct)
func TestStructDecl(t *testing.T) {
	src := `cabinet main
require "fmt"

struct User
	name String
	age Int
end

def main()
	let user User
	user.name = "Budi"
	user.age = 30
	fmt.Println(user.name, user.age)
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("RED PHASE: Failed to execute Struct. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "Budi 30") {
		t.Errorf("RED PHASE: Expected to print 'Budi 30', got: '%s'", errStr)
	}
}

// Layer 13 Test: Dynamic Data Structures (Array & Map)
func TestDataStructs(t *testing.T) {
	src := `cabinet main
require "fmt"

def main()
	let arr = Array(String){"A", "B", "C"}
	let dict = Map(String, Int){"Age": 30}
	
	arr[0] = "Z"
	dict["Score"] = 100

	fmt.Println(arr[0], dict["Age"], dict["Score"])
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("RED PHASE: Failed to execute Array/Map. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "Z 30 100") {
		t.Errorf("RED PHASE: Expected to print 'Z 30 100', got: '%s'", errStr)
	}
}

// Layer 14 Test: Collective Iteration (loop & foreach)
func TestLoops(t *testing.T) {
	src := `cabinet main
require "fmt"

def main()
	let arr = Array(String){"One", "Two"}
	
	foreach idx, val in arr
		fmt.Println(idx, val)
	end
	
	let k = 0
	loop
		if k == 2
			break
		end
		fmt.Println("Loop", k)
		k = k + 1
	end
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("RED PHASE: Failed to execute loop/foreach. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "0 One") || !strings.Contains(errStr, "1 Two") || !strings.Contains(errStr, "Loop 0") || !strings.Contains(errStr, "Loop 1") {
		t.Errorf("RED PHASE: Iteration output mismatch, got: '%s'", errStr)
	}
}

// Layer 15 Test: Multi-Option Branching (case & when)
func TestSwitchCase(t *testing.T) {
	src := `cabinet main
require "fmt"

def CheckNum(x Int)
	case x
	when 1
		fmt.Println("One")
	when 2, 3
		fmt.Println("TwoThree")
	else
		fmt.Println("Other")
	end
end

def main()
	CheckNum(1)
	CheckNum(3)
	CheckNum(99)
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("RED PHASE: Failed to execute case/when. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "One") || !strings.Contains(errStr, "TwoThree") || !strings.Contains(errStr, "Other") {
		t.Errorf("RED PHASE: Switch-case output mismatch, got: '%s'", errStr)
	}
}

// Layer 16 Test: Error Handling (Blood Lock Metaprogramming)
func TestErrorHandling(t *testing.T) {
	src := `cabinet main
require "fmt"
require "errors"

def CheckError(trigger Int)
	let err = errors.New("Danger!")
	
	if trigger == 1
		err = nil
	end

	if err
		fmt.Println("Failed:", err.Error())
	end
	
	if !err
		fmt.Println("Success!")
	end
end

def main()
	CheckError(0)
	CheckError(1)
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("RED PHASE: Failed to execute error handling. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "Failed: Danger!") || !strings.Contains(errStr, "Success!") {
		t.Errorf("RED PHASE: If err output mismatch, got: '%s'", errStr)
	}
}

// Layer 17 Test: Concurrency & Memory (go, defer, Channel)
func TestConcurrency(t *testing.T) {
	src := `cabinet main
require "fmt"
require "time"

def Worker(id Int, c Channel(String))
	time.Sleep(10 * time.Millisecond)
	c <- "Task Done"
end

def main()
	let c = make(Channel(String))
	defer fmt.Println("Defer Executed")
	
	go Worker(1, c)
	
	let result = <-c
	fmt.Println("Message:", result)
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("RED PHASE: Failed to execute concurrency. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	// Because defer is executed at the end (after function finishes), message prints before/after depending on flush, 
	// but in pure Go standard output will capture both.
	if !strings.Contains(errStr, "Message: Task Done") || !strings.Contains(errStr, "Defer Executed") {
		t.Errorf("RED PHASE: Concurrency output incomplete, got: '%s'", errStr)
	}
}

// Layer 18 Test: Advanced Concurrency (Select)
func TestSelectStatement(t *testing.T) {
	src := `cabinet main
require "fmt"
require "time"

def Send1(c Channel(String))
	time.Sleep(10 * time.Millisecond)
	c <- "One"
end

def Send2(c Channel(String))
	time.Sleep(20 * time.Millisecond)
	c <- "Two"
end

def main()
	let c1 = make(Channel(String))
	let c2 = make(Channel(String))

	go Send1(c1)
	go Send2(c2)

	select
	when <-c1
		fmt.Println("Got from c1: One")
	when <-c2
		fmt.Println("Got from c2: Two")
	when <-time.After(50 * time.Millisecond)
		fmt.Println("Timeout")
	end
end`

	var out bytes.Buffer
	err := RunSource([]byte(src), &out)
	
	if err != nil {
		t.Fatalf("RED PHASE: Failed to execute select. Error: %v, out: %s", err, out.String())
	}

	errStr := out.String()
	if !strings.Contains(errStr, "Got from c1: One") {
		t.Errorf("RED PHASE: Select output mismatch, got: '%s'", errStr)
	}
}
