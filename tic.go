package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	return newTic(workDir, slug, true, true)
}

func newServerTic(workDir string, slug string) (*Tic, error) {
	return newTic(workDir, slug, true, false)
}

func newTic(workDir string, slug string, hasImportFile bool, hasExportFile bool) (*Tic, error) {
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
		args = append(args, "--delay=5")
		args = append(args, fmt.Sprintf("--codeimport=%s", tic.importFilename))
	}

	if hasExportFile {
		tic.exportFilename = filepath.Clean(fmt.Sprintf("%sexport-%s.lua", workDir, slug))
		args = append(args, fmt.Sprintf("--codeexport=%s", tic.exportFilename))
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

func (t *Tic) importCode(code []byte) error {
	return os.WriteFile(t.importFilename, code, 0644)
}

func (t *Tic) exportCode() ([]byte, error) {
	return os.ReadFile(t.exportFilename)
}
