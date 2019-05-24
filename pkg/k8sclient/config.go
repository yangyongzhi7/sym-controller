package k8sclient

import (
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// NewClientConfig creates a Kubernetes client config from raw kube config.
func NewClientConfig(kubeConfig []byte) (*rest.Config, error) {
	if kubeConfig == nil {
		return nil, errors.New("kube config is empty")
	}
	apiconfig, err := clientcmd.Load(kubeConfig)
	if err != nil {
		return nil, emperror.Wrap(err, "failed to load kubernetes API config")
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*apiconfig, &clientcmd.ConfigOverrides{})
	config, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, emperror.Wrap(err, "failed to build client config from API config")
	}

	return config, nil
}
