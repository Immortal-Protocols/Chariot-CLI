package cmd

import (
	"encoding/json"
	"net/http"
	"testing"
)

// `rename <agent> <name>` PUTs the name to the per-agent endpoint; the agent
// arg is passed through verbatim (the backend resolves id, slug, or name).
func TestRenameSetsName(t *testing.T) {
	var body map[string]any
	var path, method string
	login(t, func(w http.ResponseWriter, r *http.Request) {
		path, method = r.URL.Path, r.Method
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"slug":"agent-000003","name":"researcher"}`))
	})

	got := runCLI(t, "", "rename", "agent-000003", "researcher")
	if got.err != nil {
		t.Fatalf("rename: %v", got.err)
	}
	if method != http.MethodPut || path != "/v1/agents/agent-000003/name" {
		t.Errorf("request = %s %s, want PUT /v1/agents/agent-000003/name", method, path)
	}
	if body["name"] != "researcher" {
		t.Errorf("name = %v", body["name"])
	}
	mustContain(t, got.stdout, "✓ agent-000003 is now named researcher", "stdout")
}

// `rename <agent> --clear` must send a JSON null (not an empty string) so the
// backend clears the name.
func TestRenameClearSendsNull(t *testing.T) {
	var body map[string]any
	login(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&body)
		_, _ = w.Write([]byte(`{"slug":"agent-000003","name":null}`))
	})

	got := runCLI(t, "", "rename", "researcher", "--clear")
	if got.err != nil {
		t.Fatalf("rename --clear: %v", got.err)
	}
	raw, present := body["name"]
	if !present {
		t.Fatal("body must carry an explicit name key")
	}
	if raw != nil {
		t.Errorf("name = %v, want JSON null", raw)
	}
	mustContain(t, got.stdout, "✓ agent-000003 is now unnamed", "stdout")
}

func TestRenameArgValidation(t *testing.T) {
	logout(t)
	if got := runCLI(t, "", "rename", "agent-000003"); got.err == nil {
		t.Error("rename without a name must error")
	}
	if got := runCLI(t, "", "rename", "a", "b", "--clear"); got.err == nil {
		t.Error("rename --clear with a name must error")
	}
}
