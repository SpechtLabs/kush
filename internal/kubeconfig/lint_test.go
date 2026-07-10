package kubeconfig

import (
	"strings"
	"testing"

	"k8s.io/client-go/tools/clientcmd/api"
)

func TestLintMissingClusterAndUser(t *testing.T) {
	cfg := api.NewConfig()
	cfg.Contexts["broken"] = &api.Context{Cluster: "ghost", AuthInfo: "nobody"}
	cfg.CurrentContext = "broken"

	findings := Lint(cfg, nil)
	joined := renderFindings(findings)
	if !strings.Contains(joined, "cluster") || !strings.Contains(joined, "ghost") {
		t.Fatalf("expected missing-cluster finding, got: %s", joined)
	}
	if !strings.Contains(joined, "user") || !strings.Contains(joined, "nobody") {
		t.Fatalf("expected missing-user finding, got: %s", joined)
	}
}

func TestLintEmptyCurrentContext(t *testing.T) {
	cfg := sampleConfig()
	cfg.CurrentContext = ""
	findings := Lint(cfg, nil)
	if !strings.Contains(renderFindings(findings), "current-context") {
		t.Fatalf("expected empty current-context finding, got: %s", renderFindings(findings))
	}
}

func TestLintUnreachableFile(t *testing.T) {
	findings := Lint(sampleConfig(), []string{"/no/such/kubeconfig/file"})
	if !strings.Contains(renderFindings(findings), "/no/such/kubeconfig/file") {
		t.Fatalf("expected unreachable-file finding, got: %s", renderFindings(findings))
	}
}

func TestLintClean(t *testing.T) {
	if findings := Lint(sampleConfig(), nil); len(findings) != 0 {
		t.Fatalf("clean config produced findings: %s", renderFindings(findings))
	}
}

func renderFindings(fs []Finding) string {
	var b strings.Builder
	for _, f := range fs {
		b.WriteString(f.Level)
		b.WriteString(": ")
		b.WriteString(f.Message)
		b.WriteString("\n")
	}
	return b.String()
}
