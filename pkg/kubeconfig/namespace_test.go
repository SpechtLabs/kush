package kubeconfig

import (
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

func emptyConfig() *api.Config { return api.NewConfig() }

func TestSetNamespace(t *testing.T) {
	out, err := Extract(sampleConfig(), "prod", "")
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/prod.yaml"
	if err := clientcmd.WriteToFile(*out, path); err != nil {
		t.Fatal(err)
	}

	if err := SetNamespace(path, "database"); err != nil {
		t.Fatalf("SetNamespace() error = %v", err)
	}

	reloaded, err := clientcmd.LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := reloaded.Contexts[reloaded.CurrentContext].Namespace; got != "database" {
		t.Fatalf("namespace = %q, want database", got)
	}
}

func TestSetNamespaceNoCurrentContext(t *testing.T) {
	base := t.TempDir()
	path := base + "/empty.yaml"
	if err := clientcmd.WriteToFile(*emptyConfig(), path); err != nil {
		t.Fatal(err)
	}
	if err := SetNamespace(path, "x"); err == nil {
		t.Fatal("SetNamespace() error = nil for no current-context, want error")
	}
}
