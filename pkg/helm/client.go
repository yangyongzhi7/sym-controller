package helm

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/proto/hapi/chart"
	rls "k8s.io/helm/pkg/proto/hapi/services"
	"os"
	"strings"
)

// Client encapsulates a Helm Client and a Tunnel for that client to interact with the Tiller pod
type Client struct {
	*kube.Tunnel
	*helm.Client
}

// ReleaseNotFoundError is returned when a Helm related operation is executed on
// a release (helm release) that doesn't exists
type ReleaseNotFoundError struct {
	HelmError error
}

func (e *ReleaseNotFoundError) Error() string {
	return fmt.Sprintf("release not found: %s", e.HelmError)
}

// NewClient
func NewClient(cfg *restclient.Config, kubeClient *kubernetes.Clientset) (*Client, error) {
	glog.Info("create kubernetes tunnel")
	tillerTunnel, err := portforwarder.New("kube-system", kubeClient, cfg)
	if err != nil {
		return nil, emperror.Wrap(err, "failed to create kubernetes tunnel")
	}

	tillerTunnelAddress := fmt.Sprintf("localhost:%d", tillerTunnel.Local)
	glog.Infof("created kubernetes tunnel on address:%s", tillerTunnelAddress)

	hClient := helm.NewClient(helm.Host(tillerTunnelAddress))

	return &Client{Tunnel: tillerTunnel, Client: hClient}, nil
}

//
func SaveChartByte(c *chart.Chart) ([]byte, error) {
	filename, err := chartutil.Save(c, "/tmp")
	if err != nil {
		glog.Infof("err:%#v", err)
		return nil, errors.Wrap(err, "save tmp chart fail")
	}
	defer os.Remove(filename)

	chartByte, err := ioutil.ReadFile(filename)
	if err != nil {
		glog.Infof("err:%#v", err)
		return nil, errors.Wrap(err, "read tmp chart to byte fail")
	}

	glog.Infof("name:%s, filename:%s", c.GetMetadata().Name, filename)
	return chartByte, nil
}

// GetReleaseByVersion returns the details of a helm release version
func (helmClient *Client) GetReleaseByVersion(releaseName string, version int32) (*rls.GetReleaseContentResponse, error) {
	rlsInfo, err := helmClient.ReleaseContent(releaseName, helm.ContentReleaseVersion(version))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, &ReleaseNotFoundError{HelmError: err}
		}
		return nil, err
	}

	return rlsInfo, nil
}
