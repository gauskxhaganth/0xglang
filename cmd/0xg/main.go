package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"orez/internal/compiler"
	"orez/internal/mod"
	"orez/internal/reporter"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(0)
	}

	command := os.Args[1]
	if command == "mod" {
		fmt.Println("0xg: 'mod' command is not applicable in 0xg, use '0xg project'")
		os.Exit(1)
	}
	
	args := os.Args[2:]

	// Transparent Proxy Architecture
	if strings.HasSuffix(command, ".0xg") {
		// If the user invokes `./0xg neon.0xg`, it automatically becomes `run neon.0xg`
		args = append([]string{command}, args...)
		command = "run"
	}

	var sourceFile string
	var sourceArgs []string
	var outPath string
	
	// Scan args for .0xg file and -o flag
	for i := 0; i < len(args); i++ {
		if args[i] == "-o" && i+1 < len(args) {
			outPath = args[i+1]
			if !filepath.IsAbs(outPath) {
				cwd, _ := os.Getwd()
				outPath = filepath.Join(cwd, outPath)
			}
			i++ // skip the value
			continue
		}
		if command == "env" && strings.HasPrefix(args[i], "0XG") {
			args[i] = "GO" + args[i][3:]
		}
		if strings.HasPrefix(args[i], "-") {
			sourceArgs = append(sourceArgs, args[i])
		} else if strings.HasSuffix(args[i], ".0xg") && sourceFile == "" {
			sourceFile = args[i]
		} else {
			sourceArgs = append(sourceArgs, args[i])
		}
	}

	if command == "version" {
		cmd := exec.Command("go", "version")
		out, err := cmd.Output()
		if err == nil {
			// e.g., go version go1.20 linux/amd64
			parts := strings.Split(strings.TrimSpace(string(out)), " ")
			if len(parts) >= 4 {
				goVer := parts[2]
				osArch := parts[3]
				fmt.Printf("\033[36mOrez v0.0.3\033[0m\n")
				fmt.Printf("\033[31mMachine: %s\033[0m\n", goVer)
				fmt.Printf("Use-on: %s\n", osArch)
				return
			}
		}
		fmt.Printf("\033[36mOrez v0.0.3\033[0m\n")
		return
	}

	if command == "run" && sourceFile != "" {
		src, err := os.ReadFile(sourceFile)
		if err != nil {
			fmt.Printf("Failed to read %s: %v\n", sourceFile, err)
			os.Exit(1)
		}

		outWriter := &interceptorWriter{w: os.Stdout}
		err = compiler.RunSourceFile(sourceFile, src, outWriter, sourceArgs...)
		if err != nil {
			if !strings.Contains(err.Error(), "exit status") {
				reporter.PrettyPrintError(err)
			}
			os.Exit(1)
		}
		return
	}

	if command == "build" && sourceFile != "" {
		src, err := os.ReadFile(sourceFile)
		if err != nil {
			fmt.Printf("Failed to read %s: %v\n", sourceFile, err)
			os.Exit(1)
		}

		if outPath == "" {
			cwd, _ := os.Getwd()
			binaryName := strings.TrimSuffix(filepath.Base(sourceFile), ".0xg")
			outPath = filepath.Join(cwd, binaryName)
		}

		outWriter := &interceptorWriter{w: os.Stdout}
		err = compiler.BuildSourceFile(sourceFile, src, outWriter, outPath, sourceArgs)
		if err != nil {
			if !strings.Contains(err.Error(), "exit status") {
				reporter.PrettyPrintError(err)
			}
			os.Exit(1)
		}
		return
	}

	// Package Manager Intercepts
	if command == "project" && len(args) > 0 && args[0] == "init" {
		moduleName := "0xg_project"
		if len(args) > 1 {
			moduleName = args[1]
		}
		if err := mod.InitMod(moduleName); err != nil {
			fmt.Printf("Failed to initialize module: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("0xg: initialized project %s in 0xg.toml\n", moduleName)
		return
	}

	if command == "get" || (command == "project" && len(args) > 0 && args[0] == "tidy") {
		outWriter := &interceptorWriter{w: os.Stdout}
		errWriter := &interceptorWriter{w: os.Stderr}
		
		goCmd := command
		if command == "project" {
			goCmd = "mod"
		}
		modArgs := append([]string{goCmd}, args...)
		if err := mod.RunModCommand(modArgs, outWriter, errWriter); err != nil {
			// errors are already caught within RunModCommand if stdout leaks
			os.Exit(1)
		}
		return
	}

	if command == "work" {
		outWriter := &interceptorWriter{w: os.Stdout}
		errWriter := &interceptorWriter{w: os.Stderr}
		
		if b, err := os.ReadFile("0xg.work"); err == nil {
			os.WriteFile("go.work", b, 0644)
		}
		if b, err := os.ReadFile("0xg.work.lock"); err == nil {
			os.WriteFile("go.work.sum", b, 0644)
		}

		cmd := exec.Command("go", append([]string{"work"}, args...)...)
		cmd.Stdout = outWriter
		cmd.Stderr = errWriter
		cmd.Stdin = os.Stdin
		cmd.Run()
		
		if b, err := os.ReadFile("go.work"); err == nil {
			os.WriteFile("0xg.work", b, 0644)
			os.Remove("go.work")
		}
		if b, err := os.ReadFile("go.work.sum"); err == nil {
			os.WriteFile("0xg.work.lock", b, 0644)
			os.Remove("go.work.sum")
		}
		os.Exit(0)
	}

	outWriter := &interceptorWriter{w: os.Stdout}
	errWriter := &interceptorWriter{w: os.Stderr}

	// Workspace Commands
	workspaceCommands := map[string]bool{
		"build": true, "run": true, "test": true, "fmt": true,
		"vet": true, "doc": true, "install": true, "list": true,
		"fix": true, "generate": true,
	}

	if workspaceCommands[command] {
		err := compiler.ExecWorkspace(command, args, outWriter, errWriter)
		if err != nil {
			if !strings.HasPrefix(err.Error(), "command failed") { // only print compiler errors, go command errors are printed by go itself
				reporter.PrettyPrintError(err)
			}
			os.Exit(1)
		}
		return
	}

	goCmd := command
	if goCmd == "project" {
		goCmd = "mod"
	}
	if goCmd == "help" && len(args) > 0 {
		if args[0] == "project" {
			args[0] = "mod"
		} else if args[0] == "projects" {
			args[0] = "modules"
		} else if args[0] == "project-auth" {
			args[0] = "module-auth"
		}
	}
	cmd := exec.Command("go", append([]string{goCmd}, args...)...)
	
	cmd.Stdout = outWriter
	cmd.Stderr = errWriter
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		// Return a silent exit code without adding garbage panic/error messages
		os.Exit(1)
	}
}

type interceptorWriter struct {
	w *os.File
}

func (iw *interceptorWriter) Write(p []byte) (n int, err error) {
	s := string(p)
	// Rebrand Go CLI to 0xg
	// Rebrand Go CLI to 0xg
	s = strings.ReplaceAll(s, "named files must be .go files", "named files must be .0xg files")
	s = strings.ReplaceAll(s, ".go file", ".0xg file")
	
	// Carefully replace 'go command' etc. without breaking regex
	s = strings.ReplaceAll(s, "go build", "0xg build")
	s = strings.ReplaceAll(s, "go test", "0xg test")
	s = strings.ReplaceAll(s, "go run", "0xg run")
	s = strings.ReplaceAll(s, "go fmt", "0xg fmt")
	s = strings.ReplaceAll(s, "go vet", "0xg vet")
	s = strings.ReplaceAll(s, "go doc", "0xg doc")
	s = strings.ReplaceAll(s, "go env", "0xg env")
	s = strings.ReplaceAll(s, "Go mod", "0xg project")
	s = strings.ReplaceAll(s, "go mod", "0xg project")
	s = strings.ReplaceAll(s, "go get", "0xg get")
	s = strings.ReplaceAll(s, "go list", "0xg list")
	s = strings.ReplaceAll(s, "go work", "0xg work")
	s = strings.ReplaceAll(s, "go clean", "0xg clean")
	s = strings.ReplaceAll(s, "go install", "0xg install")
	s = strings.ReplaceAll(s, "go generate", "0xg generate")
	s = strings.ReplaceAll(s, "go tool", "0xg tool")
	s = strings.ReplaceAll(s, "go bug", "0xg bug")
	s = strings.ReplaceAll(s, "go version", "0xg version")
	s = strings.ReplaceAll(s, "go telemetry", "0xg telemetry")

	versionRe := regexp.MustCompile(`(?m)^0xg version go(1\.\d+(?:\.\d+)?) (\w+/\w+)`)
	s = versionRe.ReplaceAllString(s, "0xg version v0.1.0 (compatible with go$1) $2")

	s = strings.ReplaceAll(s, "golang.org", "0xg.org")
	s = strings.ReplaceAll(s, "GOWORK", "0XGWORK")
	s = strings.ReplaceAll(s, "GOMOD", "0XGMOD")
	s = strings.ReplaceAll(s, "GOPROXY", "0XGPROXY")
	s = strings.ReplaceAll(s, "GOSUMDB", "0XGSUMDB")
	s = strings.ReplaceAll(s, "GOPRIVATE", "0XGPRIVATE")
	s = strings.ReplaceAll(s, "GOMODCACHE", "0XGMODCACHE")
	s = strings.ReplaceAll(s, "GOENV", "0XGENV")
	s = strings.ReplaceAll(s, "GOROOT", "0XGROOT")
	s = strings.ReplaceAll(s, "GOTOOLDIR", "0XGTOOLDIR")
	s = strings.ReplaceAll(s, "GOVERSION", "0XGVERSION")
	s = strings.ReplaceAll(s, "GO111MODULE", "0XG111MODULE")
	s = strings.ReplaceAll(s, "GOCACHE", "0XGCACHE")
	
	envRegex := regexp.MustCompile(`(?m)^GO([A-Z0-9_]+)=`)
	s = envRegex.ReplaceAllString(s, "0XG$1=")
	
	s = strings.ReplaceAll(s, "Go team", "0xg team")
	s = strings.ReplaceAll(s, "golang", "0xg")
	
	// Help text specific replacements
	s = strings.ReplaceAll(s, "mod         module maintenance", "project     project maintenance")
	s = strings.ReplaceAll(s, "modules         modules, module versions, and more", "projects        projects, project versions, and more")
	s = strings.ReplaceAll(s, "module-auth     module authentication", "project-auth    project authentication")
	s = strings.ReplaceAll(s, "packages        package lists and patterns", "cabinets        cabinet lists and patterns")
	s = strings.ReplaceAll(s, "packages        cabinet lists and patterns", "cabinets        cabinet lists and patterns")
	
	s = strings.ReplaceAll(s, "Modules", "Projects")
	s = strings.ReplaceAll(s, "modules", "projects")
	s = strings.ReplaceAll(s, "Module", "Project")
	s = strings.ReplaceAll(s, "module", "project")

	if strings.HasPrefix(s, "go ") {
		s = "0xg " + s[3:]
	}
	if strings.HasPrefix(s, "go: ") {
		s = "0xg: " + s[4:]
	}
	s = strings.ReplaceAll(s, "\ngo: ", "\n0xg: ")
	s = strings.ReplaceAll(s, "\ngo ", "\n0xg ")
	s = strings.ReplaceAll(s, "Go ", "0xg ")
	s = strings.ReplaceAll(s, " go ", " 0xg ")
	s = strings.ReplaceAll(s, "\tgo ", "\t0xg ")
	s = strings.ReplaceAll(s, "'go ", "'0xg ")
	s = strings.ReplaceAll(s, "\"go ", "\"0xg ")
	s = strings.ReplaceAll(s, "go <command>", "0xg <command>")
	s = strings.ReplaceAll(s, "go help", "0xg help")
	
	// Issue 5: Hide "package ... is not in std" and remove Go path leaks
	if start := strings.Index(s, "is not in std ("); start != -1 {
		if end := strings.Index(s[start:], ")"); end != -1 {
			s = s[:start] + "is not found in 0xg standard libraries" + s[start+end+1:]
		}
	}
	s = strings.ReplaceAll(s, "go.mod", "0xg.toml")
	s = strings.ReplaceAll(s, "go.sum", "0xg.lock")
	s = strings.ReplaceAll(s, "go.work", "0xg.work")
	s = strings.ReplaceAll(s, "Go's", "0xg's")
	s = strings.ReplaceAll(s, "go.dev", "0xg.org")
	
	// Help text specific replacements
	s = strings.ReplaceAll(s, "packages", "cabinets")
	s = strings.ReplaceAll(s, "package", "cabinet") // Will catch "cabinet " too, let's be careful
	s = strings.ReplaceAll(s, "cabinet ", "cabinet ") // no-op, just in case
	
	s = strings.ReplaceAll(s, "0xg help mod ", "0xg help project ")
	s = strings.ReplaceAll(s, "go help mod ", "0xg help project ")
	s = strings.ReplaceAll(s, "module proxy", "project proxy")
	s = strings.ReplaceAll(s, "module requirements", "project requirements")
	s = strings.ReplaceAll(s, "importpath", "requirepath")
	s = strings.ReplaceAll(s, "import path", "require path")
	s = strings.ReplaceAll(s, "gofmt", "0xgfmt")
	s = strings.ReplaceAll(s, "GOPATH", "0XGPATH")
	s = strings.ReplaceAll(s, "gopath", "0xgpath")
	s = strings.ReplaceAll(s, "GOAUTH", "0XGAUTH")
	s = strings.ReplaceAll(s, "goauth", "0xgauth")
	s = strings.ReplaceAll(s, "GOVCS", "0XGVCS")
	s = strings.ReplaceAll(s, "goproxy", "0xgproxy")
	s = strings.ReplaceAll(s, "Package ", "Cabinet ")
	s = strings.ReplaceAll(s, "package ", "cabinet ")
	s = strings.ReplaceAll(s, ".go\n", ".0xg\n")
	s = strings.ReplaceAll(s, "# my-app\n", "")
	s = strings.ReplaceAll(s, "# command-line-arguments\n", "")
	s = strings.ReplaceAll(s, "no new variables on left side of :=", "FATAL: Do not use 'let' for variable reassignment.")
	s = strings.ReplaceAll(s, "declared and not used:", "unused variable:")
	
	// Custom 0xg Crash Report for Runtime Panic
	s = strings.ReplaceAll(s, "panic: runtime error:", "================ 0xg Crash Report ================\nFatal Error:")
	s = strings.ReplaceAll(s, "panic:", "================ 0xg Crash Report ================\nFatal Error:")
	s = strings.ReplaceAll(s, "goroutine ", "fiber ")

	// Intercept Go compiler errors that match: filename.0xg:line:col: message
	// Or filename.0xg:line: message
	goErrRe := regexp.MustCompile(`(?m)^(.*(?:\.0xg|\.go)):(\d+):(?:(\d+):)?\s*(.*)$`)
	if goErrRe.MatchString(s) {
		lines := strings.Split(s, "\n")
		var finalOut strings.Builder
		for _, l := range lines {
			if matches := goErrRe.FindStringSubmatch(l); matches != nil {
				filename := matches[1]
				if strings.HasSuffix(filename, ".go") {
					filename = "[0xg Core]"
				}
				var line, col int
				fmt.Sscanf(matches[2], "%d", &line)
				if matches[3] != "" {
					fmt.Sscanf(matches[3], "%d", &col)
				}
				reporter.PrintSingleError(&finalOut, filename, line, col, matches[4])
			} else if l != "" && !strings.HasPrefix(l, "# ") {
				finalOut.WriteString(l + "\n")
			}
		}
		s = finalOut.String()
	}

	_, err = iw.w.Write([]byte(s))
	// Always return len(p) so cmd.Run does not assume a 'short write' occurred
	return len(p), err
}

func printHelp() {
	fmt.Println("0xg is a transparent proxy and transpiler for Go (Orez).")
	fmt.Println("\nUsage:")
	fmt.Println("\t0xg <command> [arguments]")
	fmt.Println("\n0xg fully inherits and wraps the standard Go toolchain.")
	fmt.Println("For a full list of commands, run: 0xg help")
}
