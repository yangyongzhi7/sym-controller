package k8sclient

import (
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	crdclient "github.com/yangyongzhi/sym-operator/pkg/client/clientset/versioned"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// NewClient creates a new Kubernetes client from config.
func NewClientFromConfig(config *rest.Config) (*kubernetes.Clientset, error) {
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, emperror.Wrap(err, "failed to create client for config")
	}

	return client, nil
}

// NewClientFromKubeConfig creates a new Kubernetes client from raw kube config.
func NewClientFromKubeConfig(kubeConfig []byte) (*kubernetes.Clientset, error) {
	config, err := NewClientConfig(kubeConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create client config")
	}

	return NewClientFromConfig(config)
}

// NewInClusterClient returns a Kubernetes client based on in-cluster configuration.
func NewInClusterClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, emperror.Wrap(err, "failed to fetch in-cluster configuration")
	}

	return NewClientFromConfig(config)
}

// NewExtClientset
func NewExtClientset(kubeConfig []byte) (*apiextensionsclient.Clientset, error) {
	config, err := NewClientConfig(kubeConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create client config")
	}

	extClient, err := apiextensionsclient.NewForConfig(config)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create extClient")
	}

	return extClient, nil
}

// NewCrdClientset
func NewCrdClientset(kubeConfig []byte) (*crdclient.Clientset, error) {
	config, err := NewClientConfig(kubeConfig)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create client config")
	}

	crdClient, err := crdclient.NewForConfig(config)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to create extClient")
	}

	return crdClient, nil
}
