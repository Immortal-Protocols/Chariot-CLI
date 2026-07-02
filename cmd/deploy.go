package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	deployCount    int
	deployEndpoint string
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Create a fleet of agents",
	Long: `Create a fleet of agents.

Agents start deactivated and cost nothing until you message them. Deploy prints
a token-seed (shown once) — your backend uses it, together with an agent id from
` + "`chariot list`" + `, to send messages.

--endpoint is optional. With it, agents POST their replies to that webhook URL.
Without it, replies are stored server-side and you read them by polling with the
token-seed (` + "`chariot demo watch`" + `) — no public tunnel needed.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if deployCount <= 0 {
			return fmt.Errorf("--count must be positive")
		}
		client, _, err := authedClient()
		if err != nil {
			return err
		}
		res, err := client.Deploy(cmd.Context(), deployCount, deployEndpoint)
		if err != nil {
			return err
		}
		fmt.Printf("\n✓ %d agents live (total %d)\n\n", res.Created, res.Total)
		if deployEndpoint != "" {
			fmt.Printf("  endpoint    : %s\n", deployEndpoint)
		} else {
			fmt.Printf("  endpoint    : (none — replies stored server-side)\n")
		}
		fmt.Printf("  namespace   : %s\n", res.Namespace)
		fmt.Printf("  token-seed  : %s\n", res.TokenSeed)
		fmt.Println("\n  Save the token-seed now — it is shown only once.")
		fmt.Println("  Address an agent from your backend:")
		fmt.Println("    POST {endpoint-base}/v1/agents/{agent-id}/messages")
		fmt.Println("    header  X-Chariot-Token: " + res.TokenSeed)
		fmt.Println("  Find agent ids with `chariot list`.")
		if deployEndpoint != "" {
			fmt.Println("  Agent replies POST to your --endpoint.")
		} else {
			fmt.Println("  See replies with:  chariot demo watch --token " + res.TokenSeed)
		}
		return nil
	},
}

func init() {
	deployCmd.Flags().IntVar(&deployCount, "count", 0, "number of agents to create")
	deployCmd.Flags().StringVar(&deployEndpoint, "endpoint", "", "webhook URL your agents reply to (optional; omit to poll with `chariot demo watch`)")
	rootCmd.AddCommand(deployCmd)
}
