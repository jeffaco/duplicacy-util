package main

// Notifier interface exposes all notification triggers
type Notifier interface {
	NotifyOfStart() error
	NotifyOfSkip() error
	NotifyOfSuccess() error
	NotifyOfFailure() error
}
