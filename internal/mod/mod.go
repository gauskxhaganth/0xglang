package mod

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
	"golang.org/x/mod/modfile"
)

type Config struct {
	Module  ModuleConfig      `toml:"module"`
	Require map[string]string `toml:"require,omitempty"`
}

type ModuleConfig struct {
	Name    string `toml:"name"`
	Version string `toml:"version,omitempty"`
	Go      string `toml:"go,omitempty"`
}

// InitMod di-invoke oleh `0xg mod init <module_name>`
func InitMod(name string) error {
	cfg := Config{
		Module: ModuleConfig{
			Name: name,
			Go:   "1.21", // default Go version
		},
	}

	b, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile("0xg.toml", b, 0644)
}

// SyncGoMod translates 0xg.toml to go.mod in the target directory
func SyncGoMod(targetDir string, cwd string) error {
	tomlPath := filepath.Join(cwd, "0xg.toml")
	data, err := os.ReadFile(tomlPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Proceed gracefully if 0xg.toml is missing, simple go run works without mod
		}
		return fmt.Errorf("gagal membaca 0xg.toml: %v", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("gagal parse 0xg.toml: %v", err)
	}

	mf := new(modfile.File)
	mf.AddModuleStmt(cfg.Module.Name)
	if cfg.Module.Go != "" {
		mf.AddGoStmt(cfg.Module.Go)
	}

	for pkg, ver := range cfg.Require {
		mf.AddRequire(pkg, ver)
	}

	modData, err := mf.Format()
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath.Join(targetDir, "go.mod"), modData, 0644)
	if err != nil {
		return err
	}

	// Sinkronisasi 0xg.lock -> go.sum
	sumPath := filepath.Join(cwd, "0xg.lock")
	if sumData, err := os.ReadFile(sumPath); err == nil {
		os.WriteFile(filepath.Join(targetDir, "go.sum"), sumData, 0644)
	}

	return nil
}

// UpdateTomlFromGoMod reads go.mod updated by `go get` or `go mod tidy`
// then saves it back to 0xg.toml
func UpdateTomlFromGoMod(targetDir string, cwd string) error {
	modData, err := os.ReadFile(filepath.Join(targetDir, "go.mod"))
	if err != nil {
		return fmt.Errorf("gagal membaca go.mod hasil generasi: %v", err)
	}

	f, err := modfile.Parse("go.mod", modData, nil)
	if err != nil {
		return err
	}

	tomlPath := filepath.Join(cwd, "0xg.toml")
	var cfg Config
	
	if tomlData, err := os.ReadFile(tomlPath); err == nil {
		toml.Unmarshal(tomlData, &cfg)
	}

	if f.Module != nil {
		cfg.Module.Name = f.Module.Mod.Path
	}
	if f.Go != nil {
		cfg.Module.Go = f.Go.Version
	}

	cfg.Require = make(map[string]string)
	for _, req := range f.Require {
		cfg.Require[req.Mod.Path] = req.Mod.Version
	}

	b, err := toml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(tomlPath, b, 0644)
	if err != nil {
		return err
	}

	// Sinkronisasi go.sum -> 0xg.lock
	sumPath := filepath.Join(targetDir, "go.sum")
	if sumData, err := os.ReadFile(sumPath); err == nil {
		os.WriteFile(filepath.Join(cwd, "0xg.lock"), sumData, 0644)
	}

	return nil
}

// RunModCommand wraps `go get` or `go mod tidy` in /tmp environment
func RunModCommand(args []string, outWriter io.Writer, errWriter io.Writer) error {
	cwd, _ := os.Getwd()
	tmpDir, err := os.MkdirTemp("", "0xg_mod_*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	err = SyncGoMod(tmpDir, cwd)
	if err != nil {
		return err
	}

	// Initialize temporary 0xg.toml if missing to allow go command execution
	if _, err := os.Stat(filepath.Join(tmpDir, "go.mod")); os.IsNotExist(err) {
		cmdInit := exec.Command("go", "mod", "init", "tempmodule")
		cmdInit.Dir = tmpDir
		cmdInit.Run()
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = tmpDir
	cmd.Stdout = outWriter
	cmd.Stderr = errWriter
	cmd.Stdin = os.Stdin

	err = cmd.Run()
	if err != nil {
		return err
	}

	return UpdateTomlFromGoMod(tmpDir, cwd)
}
