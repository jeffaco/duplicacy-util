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

func TestValidConfigWithArray(t *testing.T) {
	configFile = newConfigurationFile()
	configFile.setConfig("array")
	globalStorageDirectory = "test/assets/backupConfigs/"
	err := configFile.loadConfig(true, true)
	if err != nil {
		t.Error(err)
	}
}

func TestValidConfig_NoCopySection(t *testing.T) {
	configFile = newConfigurationFile()
	configFile.setConfig("noCopySection")
	globalStorageDirectory = "test/assets/backupConfigs/"
	err := configFile.loadConfig(true, true)
	if err != nil {
		t.Error(err)
	}
}

func TestInvalidConfig_MissingStorage(t *testing.T) {
	configFile = newConfigurationFile()
	configFile.setConfig("missingStorage")
	globalStorageDirectory = "test/assets/backupConfigs/"
	err := configFile.loadConfig(true, true)
	if err == nil {
		t.Error("Invalid storage configuration error should have been returned")
	}
}

func TestInvalidConfig_MissingStorageName(t *testing.T) {
	configFile = newConfigurationFile()
	configFile.setConfig("missingStorageName")
	globalStorageDirectory = "test/assets/backupConfigs/"
	err := configFile.loadConfig(true, true)
	if err == nil {
		t.Error("Invalid storage configuration error should have been returned")
	}
}
