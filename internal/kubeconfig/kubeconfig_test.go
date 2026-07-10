package kubeconfig

import (
	"sort"
	"testing"

	"k8s.io/client-go/tools/clientcmd/api"
)

func sampleConfig() *api.Config {
	cfg := api.NewConfig()
	cfg.Clusters["prod-cluster"] = &api.Cluster{Server: "https://prod:6443"}
	cfg.Clusters["dev-cluster"] = &api.Cluster{Server: "https://dev:6443"}
	cfg.AuthInfos["prod-user"] = &api.AuthInfo{Token: "prod-token"}
	cfg.AuthInfos["dev-user"] = &api.AuthInfo{Token: "dev-token"}
	cfg.Contexts["prod"] = &api.Context{Cluster: "prod-cluster", AuthInfo: "prod-user", Namespace: "web"}
	cfg.Contexts["dev"] = &api.Context{Cluster: "dev-cluster", AuthInfo: "dev-user"}
	cfg.CurrentContext = "dev"
	return cfg
}

func TestExtractOne(t *testing.T) {
	out, err := Extract(sampleConfig(), "prod", "")
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if out.CurrentContext != "prod" {
		t.Fatalf("CurrentContext = %q, want prod", out.CurrentContext)
	}
	// Exactly one of each; no cross-context bleed.
	if len(out.Contexts) != 1 || len(out.Clusters) != 1 || len(out.AuthInfos) != 1 {
		t.Fatalf("bleed: contexts=%d clusters=%d authinfos=%d", len(out.Contexts), len(out.Clusters), len(out.AuthInfos))
	}
	if _, ok := out.Clusters["dev-cluster"]; ok {
		t.Fatal("dev-cluster leaked into extracted config")
	}
	if _, ok := out.AuthInfos["dev-user"]; ok {
		t.Fatal("dev-user leaked into extracted config")
	}
	if out.Contexts["prod"].Cluster != "prod-cluster" || out.Contexts["prod"].AuthInfo != "prod-user" {
		t.Fatal("extracted context lost its cluster/user references")
	}
	if out.Clusters["prod-cluster"].Server != "https://prod:6443" {
		t.Fatal("wrong cluster resolved")
	}
	if out.AuthInfos["prod-user"].Token != "prod-token" {
		t.Fatal("wrong user resolved")
	}
}

func TestExtractNamespaceOverride(t *testing.T) {
	out, err := Extract(sampleConfig(), "prod", "database")
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}
	if got := out.Contexts["prod"].Namespace; got != "database" {
		t.Fatalf("namespace = %q, want database", got)
	}
}

func TestExtractDefaultNamespacePreserved(t *testing.T) {
	out, _ := Extract(sampleConfig(), "prod", "")
	if got := out.Contexts["prod"].Namespace; got != "web" {
		t.Fatalf("namespace = %q, want web (preserved)", got)
	}
}

func TestExtractNoMutationOfSource(t *testing.T) {
	src := sampleConfig()
	_, _ = Extract(src, "prod", "database")
	if src.Contexts["prod"].Namespace != "web" {
		t.Fatal("Extract mutated the source config namespace")
	}
}

func TestExtractContextNotFound(t *testing.T) {
	_, err := Extract(sampleConfig(), "staging", "")
	if err == nil {
		t.Fatal("Extract() error = nil for missing context, want error")
	}
}

func TestExtractMissingCluster(t *testing.T) {
	cfg := sampleConfig()
	cfg.Contexts["broken"] = &api.Context{Cluster: "ghost", AuthInfo: "prod-user"}
	_, err := Extract(cfg, "broken", "")
	if err == nil {
		t.Fatal("Extract() error = nil for missing cluster, want error")
	}
}

func TestContexts(t *testing.T) {
	got := Contexts(sampleConfig())
	want := []string{"dev", "prod"}
	if len(got) != len(want) {
		t.Fatalf("Contexts() = %v, want %v", got, want)
	}
	if !sort.StringsAreSorted(got) {
		t.Fatalf("Contexts() not sorted: %v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Contexts()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
