package kubeconfig

import (
	humane "github.com/sierrasoftworks/humane-errors-go"
	"k8s.io/client-go/tools/clientcmd"
)

// SetNamespace edits the namespace of the current context in the kubeconfig at
// path, writing it back in place (mode preserved by WriteToFile at 0600).
func SetNamespace(path, namespace string) error {
	cfg, err := clientcmd.LoadFromFile(path)
	if err != nil {
		return humane.Wrap(err, "failed to load kubeconfig "+path, "the temp kubeconfig may have been removed; exit and re-enter the context")
	}
	ctxName := cfg.CurrentContext
	kctx, ok := cfg.Contexts[ctxName]
	if ctxName == "" || !ok {
		return humane.New("kubeconfig "+path+" has no current context to re-pin", "run `kush ns` from inside a kush shell")
	}
	kctx.Namespace = namespace
	if err := clientcmd.WriteToFile(*cfg, path); err != nil {
		return humane.Wrap(err, "failed to write kubeconfig "+path, "check write permissions on the temp kubeconfig")
	}
	return nil
}
