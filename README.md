# Chariot CLI

Deploy and manage enterprise agent fleets from your terminal.

```
chariot login                                        # authenticate (opens browser)
chariot deploy --count 10000 --endpoint https://…    # spin up a fleet
chariot list                                         # agents + their ids
chariot account                                      # credits + status
chariot demo send <agent-id> "hello"                 # message an agent
chariot demo watch                                   # poll for replies (no tunnel)
chariot demo serve                                   # receive replies via webhook
```

## Install

```bash
go install github.com/Immortal-Protocols/Chariot-CLI@latest
# or build locally:
go build -o chariot .
```

## The one user journey

1. `chariot login` — opens your browser to the Chariot site. Sign in (email code)
   and buy credits, then approve the CLI. The CLI stores a session token in
   `~/.chariot/config.json`.
2. `chariot deploy --count N [--endpoint URL]` — creates `N` agents (they start
   deactivated and cost nothing until messaged) and prints a **token-seed**
   (shown once). `--endpoint` is optional: with it, agents POST replies to that
   URL; without it, replies are stored server-side and you poll for them.
3. `chariot list` — shows each agent's id.
4. From your own backend, message an agent:
   ```
   POST {chariot-base}/v1/agents/{agent-id}/messages
   header  X-Chariot-Token: <token-seed>
   body    {"message": "…"}
   ```
   The agent replies to your `--endpoint`, or into the poll inbox if you deployed
   without one.

## Demo: the round-trip without a backend

`chariot demo` stands in for your backend on both sides of the loop.

### No tunnel needed (the headline demo)

Deploy **without** an `--endpoint` and every reply is stored server-side; poll
for it with the same token-seed you send with:

1. `chariot deploy --count N` — no `--endpoint`, so replies go to the inbox.
2. `chariot demo send <agent-id> "hello"` — message an agent (pass `--token` or
   set `CHARIOT_TOKEN_SEED`).
3. `chariot demo watch` — polls `GET /v1/replies` every 2s (`--interval` to
   change) and prints replies as they arrive. `--from-now` skips existing
   history; Ctrl-C to stop.

### Real webhook path

Deploy **with** an `--endpoint` and agents POST their replies to it:

1. `chariot demo serve` — a local webhook receiver that prints every reply
   POSTed to it (`{"agent_id", "message", "reply_to"}`). The hosted backend can
   only reach a public URL, so expose the port with a tunnel (ngrok,
   cloudflared) and use the tunnel URL as your deploy `--endpoint`.
2. `chariot demo send <agent-id> "hello"` — the reply arrives at the
   `--endpoint`, i.e. in the `demo serve` terminal.

## Configuration

| What | How |
|---|---|
| API base URL | `CHARIOT_API_URL` env, or `api_url` in `~/.chariot/config.json` (defaults to the hosted backend) |
| Session token | written by `chariot login` |

## Development

```bash
go build ./...
go vet ./...
go test ./...
```

Layout: `cmd/` (Cobra commands), `internal/api` (backend client), `internal/config`
(local config). CI runs build + vet + test on every push (`.github/workflows/ci.yml`).
