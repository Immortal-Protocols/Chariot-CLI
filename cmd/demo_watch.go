package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/Immortal-Protocols/Chariot-CLI/internal/api"
	"github.com/Immortal-Protocols/Chariot-CLI/internal/demo"
	"github.com/spf13/cobra"
)

var (
	demoWatchToken    string
	demoWatchInterval time.Duration
	demoWatchFromNow  bool
)

// demoWatchBatchLimit is the per-poll page size (the backend caps at 200).
const demoWatchBatchLimit = 200

var demoWatchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Poll the reply inbox and print agent replies (no tunnel needed)",
	Long: `Poll Chariot for your agents' replies and print each one — the no-tunnel
alternative to ` + "`chariot demo serve`" + `.

When you deploy without an --endpoint, replies are stored server-side instead of
POSTed to a webhook. This command polls ` + "`GET /v1/replies`" + ` with the token-seed
(the same credential ` + "`chariot demo send`" + ` uses — pass --token or set
CHARIOT_TOKEN_SEED) and prints replies as they arrive. Ctrl-C to stop.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		token := demoWatchToken
		if token == "" {
			token = os.Getenv("CHARIOT_TOKEN_SEED")
		}
		if token == "" {
			return fmt.Errorf("token-seed required — pass --token or set CHARIOT_TOKEN_SEED (printed once by `chariot deploy`)")
		}
		if demoWatchInterval <= 0 {
			return fmt.Errorf("--interval must be positive")
		}
		// Token-seed auth, no login needed — config is only read for the base URL.
		cfg, err := loadConfig()
		if err != nil {
			return err
		}
		client := api.New(cfg.BaseURL(), "")

		ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer stop()

		out := cmd.OutOrStdout()
		errOut := cmd.ErrOrStderr()

		cursor := int64(0)
		if demoWatchFromNow {
			// Advance past existing history without printing it.
			c, err := drainCursor(ctx, client, token, cursor)
			if err != nil && ctx.Err() == nil {
				return err
			}
			cursor = c
		}

		fmt.Fprintf(errOut, "watching for replies every %s — Ctrl-C to stop\n\n", demoWatchInterval)

		ticker := time.NewTicker(demoWatchInterval)
		defer ticker.Stop()
		for {
			next, err := pollAndPrint(ctx, client, token, cursor, out)
			if err != nil && ctx.Err() == nil {
				// Transient poll errors shouldn't kill the watcher.
				fmt.Fprintf(errOut, "poll error: %v\n", err)
			}
			cursor = next

			select {
			case <-ctx.Done():
				fmt.Fprintln(errOut, "\nstopped")
				return nil
			case <-ticker.C:
			}
		}
	},
}

// pollAndPrint drains every available page from cursor, printing each reply, and
// returns the new cursor. It stops early if the context is cancelled.
func pollAndPrint(ctx context.Context, client *api.Client, token string, cursor int64, out io.Writer) (int64, error) {
	for {
		if ctx.Err() != nil {
			return cursor, nil
		}
		page, err := client.ListReplies(ctx, token, cursor, demoWatchBatchLimit)
		if err != nil {
			return cursor, err
		}
		for _, r := range page.Replies {
			demo.PrintReply(out, replyStamp(r.CreatedAt), r.AgentID, r.ReplyTo, r.Message)
		}
		cursor = page.NextCursor
		if len(page.Replies) < demoWatchBatchLimit {
			return cursor, nil
		}
	}
}

// drainCursor advances the cursor past all existing replies without printing.
func drainCursor(ctx context.Context, client *api.Client, token string, cursor int64) (int64, error) {
	for {
		if ctx.Err() != nil {
			return cursor, nil
		}
		page, err := client.ListReplies(ctx, token, cursor, demoWatchBatchLimit)
		if err != nil {
			return cursor, err
		}
		cursor = page.NextCursor
		if len(page.Replies) < demoWatchBatchLimit {
			return cursor, nil
		}
	}
}

// replyStamp parses the reply's created_at (RFC3339) into local time, falling
// back to now if it can't be parsed.
func replyStamp(createdAt string) time.Time {
	if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
		return t.Local()
	}
	return time.Now()
}

func init() {
	demoWatchCmd.Flags().StringVar(&demoWatchToken, "token", "", "token-seed from `chariot deploy` (or CHARIOT_TOKEN_SEED)")
	demoWatchCmd.Flags().DurationVar(&demoWatchInterval, "interval", 2*time.Second, "how often to poll for new replies")
	demoWatchCmd.Flags().BoolVar(&demoWatchFromNow, "from-now", false, "skip existing replies and only print ones that arrive from now on")
	demoCmd.AddCommand(demoWatchCmd)
}
