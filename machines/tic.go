package machines

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/creativenucleus/bytejammer/config"
	"github.com/creativenucleus/bytejammer/embed"
	"github.com/creativenucleus/bytejammer/util"
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

func (t *Tic) GetProcessID() int {
	return t.cmd.Process.Pid
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

	fmt.Println(slug)

	exchangefileBasePath := fmt.Sprintf("%s_temp", config.WORK_DIR)
	err := util.EnsurePathExists(exchangefileBasePath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	exeBasePath := fmt.Sprintf("%sexecutables", config.WORK_DIR)
	err = util.EnsurePathExists(exeBasePath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	if hasImportFile {
		tic.importFullpath, err = filepath.Abs(fmt.Sprintf("%s/import-%s.lua", exchangefileBasePath, slug))
		if err != nil {
			return nil, err
		}

		args = append(args, fmt.Sprintf("--codeimport=%s", tic.importFullpath))
	}

	if hasExportFile {
		tic.exportFullpath, err = filepath.Abs(fmt.Sprintf("%s/export-%s.lua", exchangefileBasePath, slug))
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

	// #TODO: multiversion
	tic.ticFilename = filepath.Clean(fmt.Sprintf("%s/tic80-%s.exe", exeBasePath, embed.Tic80version))
	_, err = os.Stat(tic.ticFilename)
	if err != nil {
		if !os.IsNotExist(err) { // An error we won't handle
			return nil, err
		} else { // File doesn't exist - try creating it...
			err = os.WriteFile(tic.ticFilename, embed.Tic80exe, 0700)
			if err != nil {
				return nil, err
			}
		}
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

func (t Tic) WriteImportCode(ts TicState) error {
	if t.importFullpath == "" {
		log.Fatal("Tried to import code - but file is not set up")
	}

	data, err := ts.MakeDataToImport()
	if err != nil {
		return err
	}

	return os.WriteFile(t.importFullpath, data, 0644)
}

func (t Tic) ReadExportCode() (*TicState, error) {
	if t.exportFullpath == "" {
		log.Fatal("Tried to export code - but file is not set up")
	}

	data, err := os.ReadFile(t.exportFullpath)
	if err != nil {
		return nil, err
	}

	ts, err := MakeTicStateFromExportData(data)
	if err != nil {
		return nil, err
	}

	return ts, nil
}
