//go:build windows

package watch

import "strings"

// forwardsWrites is on: psmux 3.4.0's control client delivers send-keys
// reliably and answers named keys, its -H/-l forms included.
const forwardsWrites = true

// fakeReply reports the one %error psmux 3.4.0 emits without failing: its
// control client rewrites send-keys -l/-H to "send -H" and answers
// "command timed out" although the keys already landed (psmux #799, fixed
// after v3.4.0). Swallowing that reply for a send-keys command keeps the
// write on the control pipe instead of paying a process spawn per key;
// every other command's error stays real.
func fakeReply(command string, err error) bool {
	return err != nil && strings.Contains(command, "send-keys") && strings.Contains(err.Error(), "command timed out")
}
