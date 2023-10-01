package machines

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/creativenucleus/bytejammer/config"
	"github.com/creativenucleus/bytejammer/embed"
)

type Tic struct {
	cmd         *exec.Cmd
	ticFilename string
	// Add latestImport - for server to read
	importFullpath string
	// Add latestExport - for server to read
	exportFullpath string
}

func (t *Tic) GetExportFullpath() string {
	return t.exportFullpath
}

/*
	func NewNusanServerTic(slug string, broadcaster *NusanLauncher) (*Tic, error) {
		return newTic(slug, true, false, true, broadcaster)
	}
*/

func newTic(slug string, hasImportFile bool, hasExportFile bool, isServer bool /*, broadcaster *NusanLauncher*/) (*Tic, error) {
	tic := Tic{}
	args := []string{
		"--skip",
	}

	var err error
	if hasImportFile {
		tic.importFullpath, err = filepath.Abs(fmt.Sprintf("%simport-%s.lua", config.WORK_DIR, slug))
		if err != nil {
			return nil, err
		}

		args = append(args, fmt.Sprintf("--codeimport=%s", tic.importFullpath))
	}

	if hasExportFile {
		tic.exportFullpath, err = filepath.Abs(fmt.Sprintf("%sexport-%s.lua", config.WORK_DIR, slug))
		if err != nil {
			return nil, err
		}

		args = append(args, fmt.Sprintf("--codeexport=%s", tic.exportFullpath))
	}

	if isServer {
		args = append(args, "--delay=5")
		args = append(args, "--scale=2")
	}

	//	if broadcaster == nil {
	fmt.Printf("Running TIC-80 version [%s]\n", embed.Tic80version)

	tic.ticFilename = filepath.Clean(fmt.Sprintf("%stic80-%s.exe", config.WORK_DIR, slug))
	err = os.WriteFile(tic.ticFilename, embed.Tic80exe, 0700)
	if err != nil {
		return nil, err
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
	/*
		} else {
			fmt.Printf("Running broadcast TIC-80 version\n")

			(*broadcaster.ch) <- fmt.Sprintf("--codeimport=%s", filepath.Clean(tic.importFullpath))
		}
	*/

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

	if t.importFullpath != "" {
		os.Remove(t.importFullpath)
	}

	if t.exportFullpath != "" {
		os.Remove(t.exportFullpath)
	}
}

func TicCodeAddRunSignal(code []byte) []byte {
	return append([]byte("-- pos: 0,0\n"), code...)
}

func TicCodeAddAuthor(code []byte, author string) []byte {
	shim := TicCodeReplace(embed.LuaAuthorShim, map[string]string{"DISPLAY_NAME": author})
	return append(code, shim...)
}

func TicCodeReplace(code []byte, replacements map[string]string) []byte {
	args := make([]string, 0)
	for k, v := range replacements {
		key := fmt.Sprintf("--[[$%s]]--", k)
		args = append(args, key)
		args = append(args, v)
	}

	replacer := strings.NewReplacer(args...)

	return []byte(replacer.Replace(string(code)))
}

func (t *Tic) ImportCode(code []byte) error {
	if t.importFullpath == "" {
		log.Fatal("Tried to import code - but file is not set up")
	}

	return os.WriteFile(t.importFullpath, code, 0644)
}

func (t *Tic) ExportCode() ([]byte, error) {
	if t.exportFullpath == "" {
		log.Fatal("Tried to export code - but file is not set up")
	}

	return os.ReadFile(t.exportFullpath)
}
