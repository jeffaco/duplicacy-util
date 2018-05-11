package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// Location of duplicacy binary
	duplicacyPath string

	// Directory for lock files
	globalLockDir string

	// Directory for log files
	globalLogDir string
)

// loadGlobalConfig reads in config file and ENV variables if set.
func loadGlobalConfig(cfgFile string) error {
	var err error

	// Read in (or set) global environment variables
	if err = setGlobalConfigVariables(cfgFile); err != nil {
		return err
	}

	// Validate global environment variables
	if _, err = exec.LookPath(duplicacyPath); err != nil {
		fmt.Fprintln(os.Stderr, "Error", err)
		return err
	}

	if err = verifyPathExists(globalLockDir); err != nil {
		return err
	}

	os.Mkdir(globalLogDir, 0755)
	if err = verifyPathExists(globalLogDir); err != nil {
		return err
	}

	return nil
}

// Read configuration file or set reasonable defaults if none
func setGlobalConfigVariables(cfgFile string) error {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error", err)
		return err
	}

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name "duplicacy-util" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath("$HOME/.duplicacy-util")
		viper.SetConfigName("duplicacy-util")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// Set some defaults that we can depend on
	duplicacyPath = "duplicacy"
	globalLockDir = filepath.Join(home, ".duplicacy-util")
	globalLogDir = filepath.Join(home, ".duplicacy-util", "log")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// No configuration file is okay unless we specifically asked for a named file
		if cfgFile != "" {
			fmt.Fprintln(os.Stdout, "Error:", err)
			return err
		}
		return nil
	}

	fmt.Println("Using global config:", viper.ConfigFileUsed())

	var configStr string

	configStr = viper.GetString("duplicacypath")
	if configStr != "" {
		duplicacyPath = configStr
	}

	configStr = viper.GetString("lockdirectory")
	if configStr != "" {
		globalLockDir = configStr
	}

	configStr = viper.GetString("logdirectory")
	if configStr != "" {
		globalLogDir = configStr
	}

	return err
}

func verifyPathExists(path string) error {
	var err error

	if _, err = os.Stat(path); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
		return err
	}

	return nil
}