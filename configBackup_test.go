package main

import (
	"testing"
)

func TestValidConfigWithNumberedKeys(t *testing.T) {
	configFile = newConfigurationFile()
	configFile.setConfig("numberedKeys")
	globalStorageDirectory = "test/assets/backupConfigs/"
	err := configFile.loadConfig(true, true)
	if err != nil {
		t.Error(err)
	}
}
