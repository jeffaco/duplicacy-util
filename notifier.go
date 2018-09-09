package main

// Notifier interface exposes all notification triggers
type Notifier interface {
	NotifyOfStart() error
	NotifyOfSuccess() error
	NotifyOfFailure() error
}
