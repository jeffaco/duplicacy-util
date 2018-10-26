package main

import (
	"errors"
)

func notifyOfStart() error {
	var savedError error
	for _, notifier := range onStartNotifiers {
		if err := notifier.NotifyOfStart(); err != nil {
			savedError = err
		}
	}

	return savedError
}

func notifyOfSkip() error {
	var savedError error
	for _, notifier := range onSkipNotifiers {
		if err := notifier.NotifyOfSkip(); err != nil {
			savedError = err
		}
	}

	return savedError
}

func notifyOfSuccess() error {
	var savedError error
	for _, notifier := range onSuccessNotifiers {
		if err := notifier.NotifyOfSuccess(); err != nil {
			savedError = err
		}
	}

	return savedError
}

func notifyOfFailure() error {
	var savedError error
	for _, notifier := range onFailureNotifiers {
		if err := notifier.NotifyOfFailure(); err != nil {
			savedError = err
		}
	}

	return savedError
}

func testNotifications() error {
	var savedError error
	cmdConfig = "test"

	backupTable = []backupRevision{
		{
			storage:          "b2",
			chunkTotalCount:  "149",
			chunkTotalSize:   "870,624K",
			filesTotalCount:  "345",
			filesTotalSize:   "823,261K",
			filesNewCount:    "1",
			filesNewSize:     "7,984K",
			chunkNewCount:    "6",
			chunkNewSize:     "8,106K",
			chunkNewUploaded: "3,410K",
			duration:         "9 seconds",
		},
		{
			storage:          "azure-direct",
			chunkTotalCount:  "149",
			chunkTotalSize:   "870,624K",
			filesTotalCount:  "345",
			filesTotalSize:   "823,261K",
			filesNewCount:    "1",
			filesNewSize:     "7,984K",
			chunkNewCount:    "6",
			chunkNewSize:     "8,106K",
			chunkNewUploaded: "3,410K",
			duration:         "2 seconds",
		},
	}

	copyTable = []copyRevision{
		{
			storageFrom:     "b2",
			storageTo:       "azure-direct",
			chunkTotalCount: "109",
			chunkCopyCount:  "3",
			chunkSkipCount:  "106",
			duration:        "9 seconds",
		},
	}

	// Testing notifications while no notifications are set makes no sense
	if len(onFailureNotifiers) == 0 {
		return errors.New("Warning: No notifiers are configured")
	}

	if err := notifyOfStart(); err != nil {
		savedError = err
	}

	if err := notifyOfSkip(); err != nil {
		savedError = err
	}

	if err := notifyOfSuccess(); err != nil {
		savedError = err
	}

	if err := notifyOfFailure(); err != nil {
		savedError = err
	}

	return savedError
}
