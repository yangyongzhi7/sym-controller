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
	"os"
	"strings"
)

// Client encapsulates a Helm Client and a Tunnel for that client to interact with the Tiller pod
type Client struct {
	*kube.Tunnel
	*helm.Client
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

func (helmClient *Client) Install(releaseName string) {
	if len(strings.TrimSpace(releaseName)) == 0 {
		return
	}

	if namespace == "" {
		glog.Warning("Deployment namespace was not set failing back to default")
		namespace = DefaultNamespace
	}

	basicOptions := []helm.InstallOption{
		helm.ReleaseName(releaseName),
		helm.InstallDryRun(dryRun),
	}
	installOptions := append(DefaultInstallOptions, basicOptions...)
	installOptions = append(installOptions, overrideOpts...)

	installRes, err := helmClient.Client.InstallReleaseFromChart(
		chartRequested,
		namespace,
		installOptions...,
	)
}
