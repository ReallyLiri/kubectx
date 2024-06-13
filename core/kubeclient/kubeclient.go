package kubeclient

import (
	"context"
	"github.com/ahmetb/kubectx/core/kubeconfig"
	"github.com/pkg/errors"
	errors2 "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
)

func QueryNamespaces(kc *kubeconfig.Kubeconfig) ([]string, error) {
	if os.Getenv("_MOCK_NAMESPACES") != "" {
		return []string{"ns1", "ns2"}, nil
	}

	clientset, err := newKubernetesClientSet(kc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize k8s REST client")
	}

	var out []string
	var next string
	for {
		list, err := clientset.CoreV1().Namespaces().List(
			context.Background(),
			metav1.ListOptions{
				Limit:    500,
				Continue: next,
			})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list namespaces from k8s API")
		}
		next = list.Continue
		for _, it := range list.Items {
			out = append(out, it.Name)
		}
		if next == "" {
			break
		}
	}
	return out, nil
}

func NamespaceExists(kc *kubeconfig.Kubeconfig, ns string) (bool, error) {
	// for tests
	if os.Getenv("_MOCK_NAMESPACES") != "" {
		return ns == "ns1" || ns == "ns2", nil
	}

	clientset, err := newKubernetesClientSet(kc)
	if err != nil {
		return false, errors.Wrap(err, "failed to initialize k8s REST client")
	}

	namespace, err := clientset.CoreV1().Namespaces().Get(context.Background(), ns, metav1.GetOptions{})
	if errors2.IsNotFound(err) {
		return false, nil
	}
	return namespace != nil, errors.Wrapf(err, "failed to query "+
		"namespace %q from k8s API", ns)
}

func newKubernetesClientSet(kc *kubeconfig.Kubeconfig) (*kubernetes.Clientset, error) {
	b, err := kc.Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to convert in-memory kubeconfig to yaml")
	}
	cfg, err := clientcmd.RESTConfigFromKubeConfig(b)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize config")
	}
	return kubernetes.NewForConfig(cfg)
}
