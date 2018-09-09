package main

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

func notifyOfFailure() {
	for _, notifier := range onFailureNotifiers {
		_ = notifier.NotifyOfFailure()
	}
}
