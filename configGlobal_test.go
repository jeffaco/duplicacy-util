package main

import (
	"os"
	"testing"
)

func TestValidConfig(t *testing.T) {
	quietFlag = true
	err := loadGlobalConfig(".", "test/assets/globalConfigs/fullValidConfig.yml")
	if err != nil {
		t.Error(err)
	}
}

func TestConfigForParsingErrors(t *testing.T) {
	err := loadGlobalConfig(".", "test/assets/globalConfigs/corruptedConfig.yml")
	if err == nil {
		t.Error("Parsing error should have been returned")
	}
}

func TestInvalidDuplicacyPath(t *testing.T) {
	os.Setenv("DUPLICACYPATH", "/no/such/path")
	err := loadGlobalConfig(".", "test/assets/globalConfigs/emptyConfig.yml")
	if err == nil {
		t.Error("Invalid path error should have been returned")
	}
	os.Unsetenv("DUPLICACYPATH")
}

func TestInvalidLockDirectory(t *testing.T) {
	os.Setenv("LOCKDIRECTORY", "/no/such/path")
	err := loadGlobalConfig(".", "test/assets/globalConfigs/emptyConfig.yml")
	os.Unsetenv("LOCKDIRECTORY")
	if err == nil {
		t.Error("Invalid path error should have been returned")
	}
}

func TestInvalidLogDirectory(t *testing.T) {
	os.Setenv("LOGDIRECTORY", "/no/such/path")
	err := loadGlobalConfig(".", "test/assets/globalConfigs/emptyConfig.yml")
	os.Unsetenv("LOGDIRECTORY")
	if err == nil {
		t.Error("Invalid path error should have been returned")
	}
}

func TestInvalidConfigFilePath(t *testing.T) {
	err := loadGlobalConfig(".", "no/such/path")
	if err == nil {
		t.Error("Invalid path error should have been returned")
	}
}

func TestValidConfigurationChannel(t *testing.T) {
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
	_, err := configureNotificationChannel([]string{"emails"}, "onFailure")
	if err == nil {
		t.Error("Invalid notification channel error should have been returned")
	}
}

func TestForNoDuplicateNotifiers(t *testing.T) {
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
	sendMail = true
	err := loadGlobalConfig(".", "test/assets/globalConfigs/emptyConfig.yml")
	if err == nil {
		t.Error("Invalid E-main configuration error should have been returned")
	}
}

func TestSendMailFlagWithNoConfig(t *testing.T) {
	sendMail = true
	err := loadGlobalConfig(".", "")
	if err == nil {
		t.Error("Invalid E-main configuration error should have been returned")
	}
}

func TestSendMailFlagWithValidEmailConfig(t *testing.T) {
	sendMail = true
	err := loadGlobalConfig(".", "test/assets/globalConfigs/fullValidConfig.yml")
	if err != nil {
		t.Error(err)
	}
}

func TestValidDeprecatedEmailConfig(t *testing.T) {
	err := loadGlobalConfig(".", "test/assets/globalConfigs/deprecatedConfig.yml")
	if err != nil {
		t.Error(err)
	}
	_, err = NewEmailNotifier()
	if err != nil {
		t.Error(err)
	}
}
