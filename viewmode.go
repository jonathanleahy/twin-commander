package main

// ViewMode represents the current layout mode of the application.
type ViewMode int

const (
	ViewDualPane   ViewMode = iota // Traditional two-panel layout
	ViewHybridTree                 // Tree on left, file panel on right
)
