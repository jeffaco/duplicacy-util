package main

import (
	"reflect"
	"testing"
)

func TestValidConfigWithNumberedKeys(t *testing.T) {
	quietFlag = true
	defer func() {
		quietFlag = false
	}()
	cmdBackup = true
	cmdCopy = true
	cmdPrune = true
	cmdCheck = true

	// Read the expected configuration file under test
	configFile = newConfigurationFile()
	configFile.setConfig("numberedKeys")
	globalStorageDirectory = "test/assets/backupConfigs/"
	if err := configFile.loadConfig(false, false); err != nil {
		t.Error(err)
	}

	// Verify results of the configuration file load for backupInfo
	var backupInfo = []map[string]string{
		{"name": "b2", "threads": "10"},
		{"name": "azure-direct", "threads": "5"},
		{"name": "default-threads"},
	}

	if reflect.DeepEqual(backupInfo, configFile.backupInfo) == false {
		t.Error("backupInfo should have been equal, expected:", backupInfo, ", received:", configFile.backupInfo)
	}

	// Verify results of the configuration file load for copyInfo
	var copyInfo = []map[string]string{
		{"from": "b2", "to": "azure", "threads": "10"},
		{"from": "b2", "to": "default-threads"},
	}

	if reflect.DeepEqual(copyInfo, configFile.copyInfo) == false {
		t.Error("copyInfo should have been equal, expected:", copyInfo, ", received:", configFile.copyInfo)
	}

	// Verify results of the configuration file load for pruneInfo
	var pruneInfo = []map[string]string{
		{"storage": "b2", "keep": "-keep 0:365 -keep 30:180 -keep 7:30 -keep 1:7"},
		{"storage": "azure", "keep": "-keep 0:365 -keep 30:180 -keep 7:30 -keep 1:7"},
	}

	if reflect.DeepEqual(pruneInfo, configFile.pruneInfo) == false {
		t.Error("pruneInfo should have been equal, expected:", pruneInfo, ", received:", configFile.pruneInfo)
	}

	// Verify results of the configuration file load for checkInfo
	var checkInfo = []map[string]string{
		{"storage": "b2", "all": "true"},
		{"storage": "azure"},
	}

	if reflect.DeepEqual(checkInfo, configFile.checkInfo) == false {
		t.Error("checkInfo should have been equal, expected:", checkInfo, ", received:", configFile.checkInfo)
	}
}

func TestValidConfigWithArray(t *testing.T) {
	quietFlag = true
	defer func() {
		quietFlag = false
	}()

	// Read the expected configuration file under test
	configFile = newConfigurationFile()
	configFile.setConfig("array")
	globalStorageDirectory = "test/assets/backupConfigs/"
	if err := configFile.loadConfig(false, false); err != nil {
		t.Error(err)
	}

	// Verify results of the configuration file load for backupInfo
	var backupInfo = []map[string]string{
		{"name": "b2", "threads": "10"},
		{"name": "azure-direct", "threads": "5"},
		{"name": "default-threads"},
	}

	if reflect.DeepEqual(backupInfo, configFile.backupInfo) == false {
		t.Error("backupInfo should have been equal, expected:", backupInfo, ", received:", configFile.backupInfo)
	}

	// Verify results of the configuration file load for copyInfo
	var copyInfo = []map[string]string{
		{"from": "b2", "to": "azure", "threads": "10"},
		{"from": "b2", "to": "default-threads"},
	}

	if reflect.DeepEqual(copyInfo, configFile.copyInfo) == false {
		t.Error("copyInfo should have been equal, expected:", copyInfo, ", received:", configFile.copyInfo)
	}

	// Verify results of the configuration file load for pruneInfo
	var pruneInfo = []map[string]string{
		{"storage": "b2", "keep": "-keep 0:365 -keep 30:180 -keep 7:30 -keep 1:7"},
		{"storage": "azure", "keep": "-keep 0:365 -keep 30:180 -keep 7:30 -keep 1:7"},
	}

	if reflect.DeepEqual(pruneInfo, configFile.pruneInfo) == false {
		t.Error("pruneInfo should have been equal, expected:", pruneInfo, ", received:", configFile.pruneInfo)
	}

	// Verify results of the configuration file load for checkInfo
	var checkInfo = []map[string]string{
		{"storage": "b2", "all": "true"},
		{"storage": "azure"},
	}

	if reflect.DeepEqual(checkInfo, configFile.checkInfo) == false {
		t.Error("checkInfo should have been equal, expected:", checkInfo, ", received:", configFile.checkInfo)
	}
}

func TestValidConfig_NoCopySection(t *testing.T) {
	quietFlag = true
	defer func() {
		quietFlag = false
	}()

	// Read the expected configuration file under test
	configFile = newConfigurationFile()
	configFile.setConfig("noCopySection")
	globalStorageDirectory = "test/assets/backupConfigs/"
	if err := configFile.loadConfig(false, false); err != nil {
		t.Error(err)
	}

	// Verify results of the configuration file load for backupInfo
	var backupInfo = []map[string]string{
		{"name": "b2", "threads": "10"},
		{"name": "azure-direct", "threads": "5"},
		{"name": "default-threads"},
	}

	if reflect.DeepEqual(backupInfo, configFile.backupInfo) == false {
		t.Error("backupInfo should have been equal, expected:", backupInfo, ", received:", configFile.backupInfo)
	}

	// Verify results of the configuration file load for copyInfo (this test has no Copy section)
	var copyInfo []map[string]string

	if reflect.DeepEqual(copyInfo, configFile.copyInfo) == false {
		t.Error("copyInfo should have been equal, expected:", copyInfo, ", received:", configFile.copyInfo)
	}

	// Verify results of the configuration file load for pruneInfo
	var pruneInfo = []map[string]string{
		{"storage": "b2", "keep": "-keep 0:365 -keep 30:180 -keep 7:30 -keep 1:7"},
		{"storage": "azure", "keep": "-keep 0:365 -keep 30:180 -keep 7:30 -keep 1:7"},
	}

	if reflect.DeepEqual(pruneInfo, configFile.pruneInfo) == false {
		t.Error("pruneInfo should have been equal, expected:", pruneInfo, ", received:", configFile.pruneInfo)
	}

	// Verify results of the configuration file load for checkInfo
	var checkInfo = []map[string]string{
		{"storage": "b2", "all": "true"},
		{"storage": "azure"},
	}

	if reflect.DeepEqual(checkInfo, configFile.checkInfo) == false {
		t.Error("checkInfo should have been equal, expected:", checkInfo, ", received:", configFile.checkInfo)
	}
}

func TestInvalidConfig_MissingStorage(t *testing.T) {
	quietFlag = true
	defer func() {
		quietFlag = false
	}()

	configFile = newConfigurationFile()
	configFile.setConfig("missingStorage")
	globalStorageDirectory = "test/assets/backupConfigs/"
	if err := configFile.loadConfig(false, false); err == nil {
		t.Error("Invalid storage configuration error should have been returned")
	}
}

func TestInvalidConfig_MissingStorageName(t *testing.T) {
	quietFlag = true
	defer func() {
		quietFlag = false
	}()

	configFile = newConfigurationFile()
	configFile.setConfig("missingStorageName")
	globalStorageDirectory = "test/assets/backupConfigs/"
	if err := configFile.loadConfig(false, false); err == nil {
		t.Error("Invalid storage configuration error should have been returned")
	}
}
