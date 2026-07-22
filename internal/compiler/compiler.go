package compiler

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"zerouge/internal/codegen"
	"zerouge/internal/mod"
	"zerouge/internal/parser"
	"zerouge/internal/token"
)

// RunSource executes 0xg source code and streams STDOUT to writer.
func RunSource(src []byte, out io.Writer) error {
	return RunSourceFile("main.0xg", src, out)
}

// RunSourceFile executes 0xg file with original filename for precise error position reporting.
func RunSourceFile(filename string, src []byte, out io.Writer, args ...string) error {
	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, filename, src)
	if err != nil {
		return err
	}

	var goSource bytes.Buffer
	transpiler := codegen.NewTranspiler(fset)
	err = transpiler.Generate(&goSource, fileNode)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "0xg_*")
	if err != nil {
		return fmt.Errorf("gagal membuat direktori sementara: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cwd, _ := os.Getwd()
	mod.SyncGoMod(tmpDir, cwd)

	goFilePath := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(goFilePath, goSource.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("gagal menulis transpilasi: %v", err)
	}

	// Change execution from `go run main.go` to `go run .` if the module is initialized,
	// because module dependencies are better executed as a directory
	cmdArgs := []string{"run"}
	if _, err := os.Stat(filepath.Join(tmpDir, "go.mod")); err == nil {
		cmdArgs = append(cmdArgs, ".")
	} else {
		cmdArgs = append(cmdArgs, goFilePath)
	}
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = tmpDir
	cmd.Stdout = out
	cmd.Stderr = out 

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("execution failed: %v", err)
	}

	return nil
}

// BuildSourceFile compiles 0xg file to an executable binary
func BuildSourceFile(filename string, src []byte, out io.Writer, outputBinary string, args []string) error {
	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, filename, src)
	if err != nil {
		return err
	}

	var goSource bytes.Buffer
	transpiler := codegen.NewTranspiler(fset)
	err = transpiler.Generate(&goSource, fileNode)
	if err != nil {
		return err
	}

	tmpDir, err := os.MkdirTemp("", "0xg_*")
	if err != nil {
		return fmt.Errorf("gagal membuat direktori sementara: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cwd, _ := os.Getwd()
	mod.SyncGoMod(tmpDir, cwd)

	goFilePath := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(goFilePath, goSource.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("gagal menulis transpilasi: %v", err)
	}

	cmdArgs := []string{"build", "-o", outputBinary}
	cmdArgs = append(cmdArgs, args...)
	if _, err := os.Stat(filepath.Join(tmpDir, "go.mod")); err == nil {
		cmdArgs = append(cmdArgs, ".")
	} else {
		cmdArgs = append(cmdArgs, goFilePath)
	}

	cmd := exec.Command("go", cmdArgs...)
	cmd.Stdout = out
	cmd.Stderr = out
	cmd.Dir = tmpDir

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("build failed: %v", err)
	}
	return nil
}

// ExecWorkspace transpiles all .0xg files in the current workspace to a temporary directory and runs the given go command.
func ExecWorkspace(command string, args []string, out io.Writer, errOut io.Writer) error {
	tmpDir, err := os.MkdirTemp("", "0xg_*")
	if err != nil {
		return fmt.Errorf("gagal membuat direktori sementara: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cwd, _ := os.Getwd()
	
	err = filepath.Walk(cwd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		
		relPath, err := filepath.Rel(cwd, path)
		if err != nil {
			return nil
		}
		
		if info.IsDir() {
			name := info.Name()
			if name == ".git" || name == "target" || name == "blueprint" || name == "bloodlock_project" || name == "test_dictionary" || name == "test_reporter" {
				return filepath.SkipDir
			}
			return nil
		}
		
		destPath := filepath.Join(tmpDir, relPath)
		
		if strings.HasSuffix(info.Name(), ".0xg") {
			src, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			fset := token.NewFileSet()
			fileNode, err := parser.ParseFile(fset, path, src)
			if err != nil {
				return err
			}
			var goSource bytes.Buffer
			transpiler := codegen.NewTranspiler(fset)
			err = transpiler.Generate(&goSource, fileNode)
			if err != nil {
				return err
			}
			destPath = strings.TrimSuffix(destPath, ".0xg") + ".go"
			os.MkdirAll(filepath.Dir(destPath), 0755)
			os.WriteFile(destPath, goSource.Bytes(), 0644)
		} else if strings.HasSuffix(info.Name(), ".go") || strings.HasSuffix(info.Name(), ".mod") || strings.HasSuffix(info.Name(), ".sum") || strings.HasSuffix(info.Name(), ".toml") || strings.HasSuffix(info.Name(), ".txt") {
			b, err := os.ReadFile(path)
			if err == nil {
				os.MkdirAll(filepath.Dir(destPath), 0755)
				os.WriteFile(destPath, b, 0644)
			}
		}
		return nil
	})
	
	if err != nil {
		return err
	}

	mod.SyncGoMod(tmpDir, cwd)

	// Replace .0xg arguments with .go for the go tool
	goArgs := make([]string, 0, len(args))
	hasOutFlag := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-o" {
			hasOutFlag = true
			goArgs = append(goArgs, arg)
			if i+1 < len(args) {
				outPath := args[i+1]
				if !filepath.IsAbs(outPath) {
					outPath = filepath.Join(cwd, outPath)
				}
				goArgs = append(goArgs, outPath)
				i++
			}
			continue
		}
		if strings.HasSuffix(arg, ".0xg") {
			goArgs = append(goArgs, strings.TrimSuffix(arg, ".0xg") + ".go")
		} else {
			goArgs = append(goArgs, arg)
		}
	}
	
	if command == "build" && !hasOutFlag {
		outPath := filepath.Join(cwd, filepath.Base(cwd))
		goArgs = append([]string{"-o", outPath}, goArgs...)
	}

	cmdArgs := append([]string{command}, goArgs...)
	cmd := exec.Command("go", cmdArgs...)
	cmd.Dir = tmpDir
	cmd.Stdout = out
	cmd.Stderr = errOut

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("command failed: %v", err)
	}
	return nil
}
