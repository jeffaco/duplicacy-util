package main

import (
	"errors"
)

func notifyOfStart() {
	for _, notifier := range onStartNotifiers {
		_ = notifier.NotifyOfStart()
	}
}

func notifyOfSuccess() {
	for _, notifier := range onSuccessNotifiers {
		_ = notifier.NotifyOfSuccess()
	}
}

// notifyOfFailure takes a subject argument:
//   If zero length, subject will be chosen by default
//   If specified, will override the default subject
func notifyOfFailure(subject string) {
	for _, notifier := range onFailureNotifiers {
		_ = notifier.NotifyOfFailure(subject)
	}
}

func testNotifications() error {
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

	notifyOfStart()
	notifyOfSuccess()
	notifyOfFailure("")

	return nil
}
