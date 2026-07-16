package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var renameClear bool

var renameCmd = &cobra.Command{
	Use:   "rename <agent> <name>",
	Short: "Name (alias) an agent",
	Long: `Give an agent an owner-chosen name.

The name works anywhere an agent id or slug does — ssh, hibernate, delete,
models set --agent, demo send — and shows in the NAME column of chariot list.
The slug and id keep working unchanged; renaming touches no infrastructure.

    chariot rename agent-000003 researcher
    chariot rename researcher scout      # rename by the current name
    chariot rename scout --clear         # back to unnamed

Names are 1-63 lowercase letters, digits, or hyphens, unique within your
fleet.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if renameClear && len(args) != 1 {
			return fmt.Errorf("--clear takes just the agent, e.g. `chariot rename scout --clear`")
		}
		if !renameClear && len(args) != 2 {
			return fmt.Errorf("usage: chariot rename <agent> <name> (or `chariot rename <agent> --clear`)")
		}
		name := ""
		if !renameClear {
			name = args[1]
		}
		client, _, err := authedClient()
		if err != nil {
			return err
		}
		agent, err := client.SetAgentName(cmd.Context(), args[0], name)
		if err != nil {
			return err
		}
		if agent.Name != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "✓ %s is now named %s\n", agent.Slug, *agent.Name)
			return nil
		}
		fmt.Fprintf(cmd.OutOrStdout(), "✓ %s is now unnamed\n", agent.Slug)
		return nil
	},
}

func init() {
	renameCmd.Flags().BoolVar(&renameClear, "clear", false, "remove the agent's name instead of setting one")
	rootCmd.AddCommand(renameCmd)
}
