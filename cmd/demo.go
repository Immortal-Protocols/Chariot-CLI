package cmd

import (
	"github.com/spf13/cobra"
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Try the message round-trip without writing a backend",
	Long: `Demo helpers for the message round-trip.

In production your backend messages agents and receives their replies. These
commands stand in for that backend so you can try the loop from a terminal.

No tunnel needed (the headline demo) — deploy without an --endpoint and poll:

  chariot deploy --count N                 # no --endpoint: replies are stored
  chariot demo send <agent-id> "hello"     # message an agent (needs the token-seed)
  chariot demo watch                       # poll and print replies as they arrive

Real webhook path — deploy with an --endpoint and receive POSTs:

  chariot demo serve                       # local receiver (expose via a tunnel)
  chariot demo send <agent-id> "hello"     # message an agent`,
}

func init() {
	rootCmd.AddCommand(demoCmd)
}
