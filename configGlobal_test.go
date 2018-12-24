package main

import (
	"os"
	"testing"
)

func TestValidConfig(t *testing.T) {
	quietFlag = true
	runningUnitTests = true
	defer func() {
		quietFlag = false
		runningUnitTests = false
	}()

	err := loadGlobalConfig(".", "test/assets/globalConfigs/fullValidConfig.yml")
	if err != nil {
		t.Error(err)
	}
}

func TestConfigForParsingErrors(t *testing.T) {
	quietFlag = true
	runningUnitTests = true
	defer func() {
		quietFlag = false
		runningUnitTests = false
	}()

	err := loadGlobalConfig(".", "test/assets/globalConfigs/corruptedConfig.yml")
	if err == nil {
		t.Error("Parsing error should have been returned")
	}
}

func TestEmptyConfigFile(t *testing.T) {
	runningUnitTests = true
	defer func() {
		runningUnitTests = false
	}()

	err := loadGlobalConfig(os.TempDir(), "")
	if err != nil {
		t.Error("Empty configuration file should be valid")
	}
}

func TestInvalidDuplicacyPath(t *testing.T) {
	quietFlag = true
	defer func() {
		quietFlag = false
	}()

	os.Setenv("DUPLICACYPATH", "/no/such/path")
	err := loadGlobalConfig(".", "test/assets/globalConfigs/emptyConfig.yml")
	if err == nil {
		t.Error("Invalid path error should have been returned")
	}
	os.Unsetenv("DUPLICACYPATH")
}

func TestInvalidLockDirectory(t *testing.T) {
	quietFlag = true
	runningUnitTests = true
	defer func() {
		quietFlag = false
		runningUnitTests = false
	}()

	os.Setenv("LOCKDIRECTORY", "/no/such/path")
	err := loadGlobalConfig(".", "test/assets/globalConfigs/emptyConfig.yml")
	os.Unsetenv("LOCKDIRECTORY")
	if err == nil {
		t.Error("Invalid path error should have been returned")
	}
}

func TestInvalidLogDirectory(t *testing.T) {
	quietFlag = true
	runningUnitTests = true
	defer func() {
		quietFlag = false
		runningUnitTests = false
	}()

	os.Setenv("LOGDIRECTORY", "/no/such/path")
	err := loadGlobalConfig(".", "test/assets/globalConfigs/emptyConfig.yml")
	os.Unsetenv("LOGDIRECTORY")
	if err == nil {
		t.Error("Invalid path error should have been returned")
	}
}

func TestInvalidConfigFilePath(t *testing.T) {
	runningUnitTests = true
	defer func() { runningUnitTests = false }()

	err := loadGlobalConfig(".", "no/such/path")
	if err == nil {
		t.Error("Invalid path error should have been returned")
	}
}

func TestValidConfigurationChannel(t *testing.T) {
	quietFlag = true
	runningUnitTests = true
	defer func() {
		quietFlag = false
		runningUnitTests = false
	}()

	err := loadGlobalConfig(".", "test/assets/globalConfigs/fullValidConfig.yml")
	if err != nil {
		t.Error(err)
	}
	_, err = configureNotificationChannel([]string{"email"}, "onFailure")
	if err != nil {
		t.Error(err)
	}
}

func TestInvalidConfigurationChannel(t *testing.T) {
	quietFlag = true
	runningUnitTests = true
	defer func() {
		quietFlag = false
		runningUnitTests = false
	}()

	_, err := configureNotificationChannel([]string{"emails"}, "onFailure")
	if err == nil {
		t.Error("Invalid notification channel error should have been returned")
	}
}

func TestForNoDuplicateNotifiers(t *testing.T) {
	quietFlag = true
	runningUnitTests = true
	defer func() {
		quietFlag = false
		runningUnitTests = false
	}()

	err := loadGlobalConfig(".", "test/assets/globalConfigs/fullValidConfig.yml")
	if err != nil {
		t.Error(err)
	}

	var notifiers []Notifier
	notifiers, err = configureNotificationChannel([]string{"email", "email"}, "onFailure")
	if err != nil {
		t.Error(err)
	}

	if len(notifiers) != 1 {
		t.Error("Multiple identical notifiers detected")
	}
}

func TestSendMailFlagWithInvalidEmailConfig(t *testing.T) {
	quietFlag = true
	runningUnitTests = true
	testNotificationsFlag = true
	defer func() {
		quietFlag = false
		runningUnitTests = false
		testNotificationsFlag = false
	}()

	err := loadGlobalConfig(".", "test/assets/globalConfigs/emptyConfig.yml")
	if err == nil {
		t.Error("Invalid E-main configuration error should have been returned")
	}
}

func TestSendMailFlagWithValidEmailConfig(t *testing.T) {
	quietFlag = true
	runningUnitTests = true
	os.Unsetenv("DU_EMAIL_AUTH_PASSWORD")
	defer func() {
		quietFlag = false
		runningUnitTests = false
	}()

	err := loadGlobalConfig(".", "test/assets/globalConfigs/fullValidConfig.yml")
	if err != nil {
		t.Error(err)
	}

	// YAML doesn't mention the acceptInsecureCerts; it should default to false
	if emailAcceptAnyCerts {
		t.Error("acceptInsecureCerts was not specified, yet set in code")
	}

	// Verify that the auth password is what is stored in global configuration file
	if emailAuthPassword != "gaozqlwbztypagwt" {
		t.Errorf("email.authPassword should be '%s', but instead is '%s'", "gaozqlwbztypagwt", emailAuthPassword)
	}
}

func TestAcceptInsecureCertificate(t *testing.T) {
	quietFlag = true
	runningUnitTests = true
	defer func() {
		quietFlag = false
		runningUnitTests = false
	}()

	err := loadGlobalConfig(".", "test/assets/globalConfigs/fullValidConfigInsecureCerts.yml")
	if err != nil {
		t.Error(err)
	}

	// YAML sets acceptInsecureCerts to true; verify that we detected that
	if emailAcceptAnyCerts == false {
		t.Error("acceptInsecureCerts was specified, yet not set in code")
	}
}

func TestSendMailWithEnvioronmentPassword(t *testing.T) {
	quietFlag = true
	runningUnitTests = true
	os.Setenv("DU_EMAIL_AUTH_PASSWORD", "xyzzy")
	defer func() {
		quietFlag = false
		runningUnitTests = false
		os.Unsetenv("DU_EMAIL_AUTH_PASSWORD")
	}()

	err := loadGlobalConfig(".", "test/assets/globalConfigs/fullValidConfig.yml")
	if err != nil {
		t.Error(err)
	}

	// Verify that the auth password is as per the environment variale
	if emailAuthPassword != "xyzzy" {
		t.Errorf("email.authPassword should be '%s', but instead is '%s'", "xyzzy", emailAuthPassword)
	}
}
