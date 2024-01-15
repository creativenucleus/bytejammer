package machines

import (
	"errors"
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

	latestImport   TicState // #TODO: propagate this to the server panel
	importFullpath string

	// Add latestExport - for server to read
	exportFullpath string

	// This receives nil for normal shutdown (i.e. by TIC exit, user clicking the close button etc)
	chClosedErr chan error

	codeOverride bool
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

func newTic(slug string, hasImportFile bool, hasExportFile bool, isServer bool, chClosedErr chan error /*, broadcaster *NusanLauncher*/) (*Tic, error) {
	tic := Tic{}
	args := []string{
		"--skip",
	}

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
	// This may be error or nil
	go func() {
		err = tic.cmd.Wait()
		fmt.Printf("TIC (%d) finished with error: %v", tic.cmd.Process.Pid, err)
		chClosedErr <- err
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

// This pushes the supplied code to this TIC and prevents regular import for the specified duration
// 1) Grabs the current TIC state to holding
// 2) Writes the override to the import file
func (t *Tic) SetCodeOverride(tsOverride TicState, d time.Duration) error {
	// We ought to handle when someone sets multiple concurrent overrides - for the moment, just block!
	if t.codeOverride {
		return errors.New("we already have a code override - request ignored")
	}

	err := t.WriteImportCode(tsOverride, false)
	if err != nil {
		fmt.Println(err)
		return err
	}

	t.codeOverride = true

	// Remove the override struct when the timer expires
	timer := time.NewTimer(d)
	go func() error { // (NB error ignored)
		<-timer.C
		err := t.WriteImportCode(t.latestImport, true)
		if err != nil {
			return err
		}

		t.codeOverride = false
		return nil
	}()

	return nil
}

// If we have overide code, then put the supplied update in a placeholder, ready for when the override completes
func (t *Tic) WriteImportCode(ts TicState, saveAsLatest bool) error {
	if t.importFullpath == "" {
		log.Fatal("Tried to import code - but file is not set up")
	}

	if saveAsLatest {
		t.latestImport = ts
	}

	if t.codeOverride {
		return nil
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
