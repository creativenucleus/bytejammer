package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Tic struct {
	cmd         *exec.Cmd
	ticFilename string
	// Add latestImport - for server to read
	importFilename string
	// Add latestExport - for server to read
	exportFilename string
}

func newClientTic(workDir string, slug string) (*Tic, error) {
	return newTic(workDir, slug, true, true, false)
}

func newServerTic(workDir string, slug string) (*Tic, error) {
	return newTic(workDir, slug, true, false, true)
}

func newTic(workDir string, slug string, hasImportFile bool, hasExportFile bool, isServer bool) (*Tic, error) {
	fmt.Printf("Running TIC-80 version [%s]\n", embedTic80version)

	tic := Tic{}
	tic.ticFilename = filepath.Clean(fmt.Sprintf("%stic80-%s.exe", workDir, slug))
	err := os.WriteFile(tic.ticFilename, embedTic80exe, 0700)
	if err != nil {
		return nil, err
	}

	args := []string{
		"--skip",
	}

	if hasImportFile {
		tic.importFilename = filepath.Clean(fmt.Sprintf("%simport-%s.lua", workDir, slug))
		args = append(args, fmt.Sprintf("--codeimport=%s", tic.importFilename))
	}

	if hasExportFile {
		tic.exportFilename = filepath.Clean(fmt.Sprintf("%sexport-%s.lua", workDir, slug))
		args = append(args, fmt.Sprintf("--codeexport=%s", tic.exportFilename))
	}

	if isServer {
		args = append(args, "--delay=5")
		args = append(args, "--scale=1")
	}

	tic.cmd = exec.Command(tic.ticFilename, args...)
	err = tic.cmd.Start()
	if err != nil {
		return nil, err
	}

	fmt.Printf("Started TIC (pid: %d)\n", tic.cmd.Process.Pid)

	// use goroutine waiting, manage process
	// this is important, otherwise the process becomes in S mode
	go func() {
		err = tic.cmd.Wait()
		fmt.Printf("TIC (%d) finished with error: %v", tic.cmd.Process.Pid, err)
		// #TODO: cleanup
	}()

	return &tic, nil
}

func (t *Tic) shutdown() {
	// Shutdown the running program...
	if runtime.GOOS == "windows" {
		// Windows doesn't support Interrupt
		_ = t.cmd.Process.Signal(os.Kill)
	} else {
		go func() {
			time.Sleep(2 * time.Second)
			_ = t.cmd.Process.Signal(os.Kill)
		}()
		t.cmd.Process.Signal(os.Interrupt)
	}

	// #TODO: ensure program has shut down (by PID?) before removing

	// Remove temporary files...
	os.Remove(t.ticFilename)

	if t.importFilename != "" {
		os.Remove(t.importFilename)
	}

	if t.exportFilename != "" {
		os.Remove(t.exportFilename)
	}
}

func ticCodeAddRunSignal(code []byte) []byte {
	return append([]byte("-- pos: 0,0\n"), code...)
}

func ticCodeReplace(code []byte, replacements map[string]string) []byte {
	args := make([]string, 0)
	for k, v := range replacements {
		key := fmt.Sprintf("--[[$%s]]--", k)
		args = append(args, key)
		args = append(args, v)
	}

	replacer := strings.NewReplacer(args...)

	return []byte(replacer.Replace(string(code)))
}

func (t *Tic) importCode(code []byte) error {
	if t.importFilename == "" {
		log.Fatal("Tried to import code - but file is not set up")
	}

	return os.WriteFile(t.importFilename, code, 0644)
}

func (t *Tic) exportCode() ([]byte, error) {
	if t.exportFilename == "" {
		log.Fatal("Tried to export code - but file is not set up")
	}

	return os.ReadFile(t.exportFilename)
}
