package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	proxyBitcoinPIDFile = "proxy-bitcoin.pid"
	proxyLiquidPIDFile  = "proxy-liquid.pid"
)

// startProxyProcess spawns "nigiri serve" as a detached background subprocess,
// writes its PID to <datadir>/proxy-<chain>.pid, and confirms the process is
// alive after 600ms. stdout/stderr go to <datadir>/proxy-<chain>.log.
func startProxyProcess(exePath, datadir, chain string) error {
	var listenAddr, electrsAddr, rpcAddr, pidFile string
	switch chain {
	case "bitcoin":
		listenAddr = "0.0.0.0:3000"
		electrsAddr = "localhost:30000"
		rpcAddr = "localhost:18443"
		pidFile = filepath.Join(datadir, proxyBitcoinPIDFile)
	case "liquid":
		listenAddr = "0.0.0.0:3001"
		electrsAddr = "localhost:30001"
		rpcAddr = "localhost:18884"
		pidFile = filepath.Join(datadir, proxyLiquidPIDFile)
	default:
		return fmt.Errorf("unknown chain: %s", chain)
	}

	args := []string{
		"--datadir", datadir,
		"serve",
		"--chain", chain,
		"--addr", listenAddr,
		"--electrs-addr", electrsAddr,
		"--rpc-addr", rpcAddr,
		"--rpc-cookie", "admin1:123",
	}

	logFilePath := filepath.Join(datadir, "proxy-"+chain+".log")
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open proxy log file: %w", err)
	}
	defer logFile.Close()

	cmd := exec.Command(exePath, args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	// Setsid detaches from the parent's terminal and process group so the
	// subprocess survives after the parent (nigiri start) exits.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start proxy subprocess (%s): %w", chain, err)
	}

	pid := cmd.Process.Pid
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
		cmd.Process.Kill()
		return fmt.Errorf("failed to write PID file for %s proxy: %w", chain, err)
	}

	// Verify the process didn't crash immediately.
	time.Sleep(600 * time.Millisecond)
	if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
		os.Remove(pidFile)
		return fmt.Errorf("proxy subprocess (%s) died on startup (pid %d): %w", chain, pid, err)
	}

	return nil
}

// killProxyProcess sends SIGTERM to the PID in pidFile, polls until exit (5s),
// falls back to SIGKILL, then removes the PID file. Returns nil if the file
// does not exist (proxy already stopped or never started).
func killProxyProcess(pidFile string) error {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read PID file %s: %w", pidFile, err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		os.Remove(pidFile)
		return fmt.Errorf("corrupt PID file %s: %w", pidFile, err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		os.Remove(pidFile)
		return nil
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		os.Remove(pidFile)
		return nil
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		time.Sleep(100 * time.Millisecond)
		if err := proc.Signal(syscall.Signal(0)); err != nil {
			os.Remove(pidFile)
			return nil
		}
	}

	proc.Signal(syscall.SIGKILL)
	time.Sleep(200 * time.Millisecond)
	os.Remove(pidFile)
	return nil
}
