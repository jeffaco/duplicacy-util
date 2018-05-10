package main

import (
	"fmt"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// Directory for lock files
	globalLockDir string
)

// loadGlobalConfig reads in config file and ENV variables if set.
func loadGlobalConfig(cfgFile string) error {
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
	globalLockDir = path.Join(home, ".duplicacy-util")

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

	configStr := viper.GetString("lockdirectory")
	if configStr != "" {
		globalLockDir = configStr
	}
	if _, err = os.Stat(globalLockDir); err != nil {
		fmt.Fprintln(os.Stderr, "Error: ", err)
		return err
	}

	return err
}
