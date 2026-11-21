//go:build !windows

package watch

// forwardsWrites is on where tmux's control client applies and acknowledges
// every command honestly.
const forwardsWrites = true

// fakeReply is off: tmux's %error answers are real.
func fakeReply(string, error) bool { return false }
