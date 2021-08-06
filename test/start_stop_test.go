package test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/vulpemventures/nigiri/interal/state"
	"github.com/vulpemventures/nigiri/internal/config"
)

const (
	liquid  = true
	bitcoin = false
	delete  = true
)

var (
	stopCmd = "stop"
	// deleteCmd      = append(stopCmd, "--delete")
	startCmd = "start"
	// liquidStartCmd = append(startCmd, "--liquid")
	tmpDatadir  = filepath.Join(os.TempDir(), "nigiri-tmp")
	nigiriState = state.New(filepath.Join(tmpDatadir, config.DefaultName), config.InitialState)
)

func TestStartStopLiquid(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	// Start/Stop
	testStart(t, liquid)
	testStop(t)
	// Start/Delete
	testStart(t, liquid)
	testDelete(t)
}

func TestStartStopBitcoin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping testing in short mode")
	}
	// Start/Stop
	testStart(t, bitcoin)
	testStop(t)
	// Start/Delete
	testStart(t, bitcoin)
	testDelete(t)
}

func testStart(t *testing.T, flag bool) {

	if err := testCommand("start", "", flag); err != nil {
		t.Fatal(err)
	}
	//Give some time to nigiri to be ready before calling
	time.Sleep(5 * time.Second)
	if isRunning, err := nigiriState.GetBool("running"); err != nil {
		t.Fatal(err)
	} else if !isRunning {
		t.Fatal("Nigiri should have been started but services have not been found among running containers")
	}
}

func testStop(t *testing.T) {

	if err := testCommand("stop", "", !delete); err != nil {
		t.Fatal(err)
	}
	//Give some time to nigiri to be ready before calling
	time.Sleep(5 * time.Second)
	if isRunning, err := nigiriState.GetBool("running"); err != nil {
		t.Fatal(err)
	} else if isRunning {
		t.Fatal("Nigiri should have been stopped but services have not been found among stopped containers")
	}
}

func testDelete(t *testing.T) {

	if err := testCommand("stop", "", delete); err != nil {
		t.Fatal(err)
	}
	if isRunning, err := nigiriState.GetBool("running"); err != nil {
		t.Fatal(err)
	} else if isRunning {
		t.Fatal("Nigiri should have been terminated at this point but services have been found among stopped containers")
	}
}

func testCommand(command, arg string, flag bool) error {

	cmd := exec.Command("go", "run", "./cmd/nigiri")
	env := "NIGIRI_DATADIR=" + tmpDatadir
	cmd.Env = []string{env}
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	if command == "start" {
		cmd.Args = append(cmd.Args, startCmd, fmt.Sprintf("--liquid=%t", flag))
	}
	if command == "stop" {
		cmd.Args = append(cmd.Args, stopCmd, fmt.Sprintf("--delete=%t", flag))
	}

	err := cmd.Start()
	if err != nil {
		fmt.Errorf("name: %v, args: %v, err: %v", command, arg, err.Error())
	}

	return nil
}
