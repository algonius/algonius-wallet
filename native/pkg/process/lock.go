// SPDX-License-Identifier: Apache-2.0
package process

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

const (
	// PIDFileName is the name of the PID file
	PIDFileName = "algonius-wallet-host.pid"
)

// getWalletHomeDir returns the wallet home directory, respecting environment override
func getWalletHomeDir() string {
	// Check environment variable first
	if homeDir := os.Getenv("ALGONIUS_WALLET_HOME"); homeDir != "" {
		return homeDir
	}
	return ""
}

// LockPIDFile creates a PID file to prevent multiple instances
// Returns true if successfully locked, false if another instance is running
func LockPIDFile() (bool, error) {
	return LockPIDFileWithSuffix("")
}

// LockPIDFileWithSuffix creates a PID file with an optional suffix to prevent multiple instances
// Returns true if successfully locked, false if another instance is running
func LockPIDFileWithSuffix(suffix string) (bool, error) {
	// Get wallet home directory (respects ALGONIUS_WALLET_HOME environment variable)
	pidDir := getWalletHomeDir()
	if pidDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return false, fmt.Errorf("failed to get home directory: %w", err)
		}
		pidDir = filepath.Join(homeDir, ".algonius-wallet")
	}
	
	// Create full path to PID file
	fileName := PIDFileName
	if suffix != "" {
		fileName = fmt.Sprintf("algonius-wallet-host-%s.pid", suffix)
	}
	pidFilePath := filepath.Join(pidDir, fileName)
	
	// Ensure the directory exists
	if err := os.MkdirAll(pidDir, 0700); err != nil {
		return false, fmt.Errorf("failed to create PID directory: %w", err)
	}
	
	// Check if PID file already exists
	if _, err := os.Stat(pidFilePath); err == nil {
		// PID file exists, check if process is still running
		pidBytes, err := os.ReadFile(pidFilePath)
		if err != nil {
			// If we can't read the file, remove it and continue
			_ = os.Remove(pidFilePath)
		} else {
			pid, err := strconv.Atoi(string(pidBytes))
			if err != nil {
				// Invalid PID in file, remove it and continue
				_ = os.Remove(pidFilePath)
			} else {
				// Check if process is still running
				process, err := os.FindProcess(pid)
				if err == nil {
					// On Unix systems, Signal(0) tests for existence of the process
					if err := process.Signal(syscall.Signal(0)); err == nil {
						// Process is still running
						return false, nil
					}
				}
				// Process is not running, remove stale PID file
				_ = os.Remove(pidFilePath)
			}
		}
	}
	
	// Create PID file with current process ID
	pid := os.Getpid()
	pidBytes := []byte(strconv.Itoa(pid))
	
	if err := os.WriteFile(pidFilePath, pidBytes, 0644); err != nil {
		return false, fmt.Errorf("failed to write PID file: %w", err)
	}
	
	return true, nil
}

// UnlockPIDFile removes the PID file
func UnlockPIDFile() error {
	// Get wallet home directory (respects ALGONIUS_WALLET_HOME environment variable)
	pidDir := getWalletHomeDir()
	if pidDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		pidDir = filepath.Join(homeDir, ".algonius-wallet")
	}
	
	// Remove PID file
	pidFilePath := filepath.Join(pidDir, PIDFileName)
	return os.Remove(pidFilePath)
}

// KillExistingProcess kills any existing process recorded in the PID file
func KillExistingProcess() error {
	// Get wallet home directory (respects ALGONIUS_WALLET_HOME environment variable)
	pidDir := getWalletHomeDir()
	if pidDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		pidDir = filepath.Join(homeDir, ".algonius-wallet")
	}
	
	// Check if PID file exists
	pidFilePath := filepath.Join(pidDir, PIDFileName)
	
	if _, err := os.Stat(pidFilePath); err != nil {
		// No PID file exists, nothing to do
		return nil
	}
	
	// Read PID from file
	pidBytes, err := os.ReadFile(pidFilePath)
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}
	
	pid, err := strconv.Atoi(string(pidBytes))
	if err != nil {
		// Invalid PID in file, remove it
		_ = os.Remove(pidFilePath)
		return nil
	}
	
	// Try to kill the process
	process, err := os.FindProcess(pid)
	if err != nil {
		// Process doesn't exist, remove stale PID file
		_ = os.Remove(pidFilePath)
		return nil
	}
	
	// Send SIGTERM to the process
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// If that fails, the process might have already terminated
		// Remove the PID file anyway
		_ = os.Remove(pidFilePath)
		return nil
	}
	
	// Wait a moment for the process to terminate
	// Note: On some systems, we might need to wait or check if the process actually terminated
	
	// Remove the PID file
	_ = os.Remove(pidFilePath)
	
	return nil
}