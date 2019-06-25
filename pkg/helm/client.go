package helm

import (
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"github.com/goph/emperror"
	"github.com/pkg/errors"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/helm/pkg/chartutil"
	"k8s.io/helm/pkg/helm"
	helmapi "k8s.io/helm/pkg/helm"
	"k8s.io/helm/pkg/helm/portforwarder"
	"k8s.io/helm/pkg/kube"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/proto/hapi/release"
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

// Install a release.
func (helmClient *Client) InstallRelease(namespace string, releaseName string, chartBytes []byte, raw string) (*rls.InstallReleaseResponse, error) {
	//releaseResponse.GetRelease().Chart
	requestedChart, err := chartutil.LoadArchive(bytes.NewReader(chartBytes))
	if err != nil {
		glog.Infof("Load archive when you want to install a release has an error : %s", err.Error())
		return nil, err
	} else {
		response, err := helmClient.InstallReleaseFromChart(requestedChart, namespace,
			helmapi.ReleaseName(releaseName), helmapi.ValueOverrides([]byte(raw)))
		if err != nil {
			glog.Infof("Installing a release [%s] has an error : %s", releaseName, err.Error())
			return nil, err
		}

		return response, nil
	}
}

// Update a release.
func (helmClient *Client) UpdateRelease(rlsName string, chartBytes []byte, raw string) (*rls.UpdateReleaseResponse, error) {
	requestedChart, err := chartutil.LoadArchive(bytes.NewReader(chartBytes))
	if err != nil {
		glog.Infof("Load archive when you want to update a release has an error : %s", err.Error())
		return nil, err
	} else {
		updateResponse, err := helmClient.UpdateReleaseFromChart(rlsName, requestedChart, helmapi.UpdateValueOverrides([]byte(raw)))
		if err != nil {
			glog.Infof("Updating a release [%s] has an error : %s", rlsName, err.Error())
			return nil, err
		}

		return updateResponse, nil
	}
}

// Delete a release.
func (helmClient *Client) UninstallRelease(rlsName string) (*rls.UninstallReleaseResponse, error) {
	deleteResponse, err := helmClient.DeleteRelease(rlsName, helmapi.DeletePurge(true))
	if err != nil {
		glog.Infof("Delete the release [%s] has an error : %s", rlsName, err.Error())
		return nil, err
	}

	return deleteResponse, nil
}

// GetReleaseByVersion returns the details of a helm release version.
func (helmClient *Client) GetReleaseByVersion(releaseName string, version int32) (*rls.GetReleaseContentResponse, error) {
	ops := []helmapi.ContentOption{}

	if version != 0 {
		ops = append(ops, helm.ContentReleaseVersion(version))
	}

	rlsInfo, err := helmClient.ReleaseContent(releaseName, ops...)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, &ReleaseNotFoundError{HelmError: err}
		}
		return nil, err
	}

	return rlsInfo, nil
}

// Find the release with its name, return the just only one.
func (helmClient *Client) GetRelease(releaseName string) (*release.Release, error) {
	ops := []helmapi.ReleaseListOption{
		helmapi.ReleaseListSort(int32(rls.ListSort_LAST_RELEASED)),
		helmapi.ReleaseListOrder(int32(rls.ListSort_DESC)),
		helmapi.ReleaseListStatuses([]release.Status_Code{
			release.Status_DEPLOYED,
			release.Status_FAILED,
			release.Status_DELETING,
			release.Status_PENDING_INSTALL,
			release.Status_PENDING_UPGRADE,
			release.Status_PENDING_ROLLBACK}),
		// helm.ReleaseListLimit(limit),
		// helm.ReleaseListFilter(filter),
		// helm.ReleaseListNamespace(""),
	}

	ops = append(ops, helmapi.ReleaseListFilter(releaseName))
	listResponse, err := helmClient.ListReleases(ops...)
	if err != nil {
		// if the release is not exist, we should install it.
		return nil, err
	}

	if listResponse == nil {
		glog.Infof("The list response is nil, [%s]", releaseName)
		return nil, nil
	}

	if listResponse.Releases == nil || len(listResponse.Releases) < 0 {
		glog.Infof("Can not find any release named [%s]", releaseName)
		return nil, nil
	}

	return listResponse.Releases[0], nil
}

// Filter the releases what you want to find.
func (helmClient *Client) FilterReleases(regex string) ([]*release.Release, error) {
	ops := []helmapi.ReleaseListOption{
		helmapi.ReleaseListSort(int32(rls.ListSort_LAST_RELEASED)),
		helmapi.ReleaseListOrder(int32(rls.ListSort_DESC)),
		helmapi.ReleaseListStatuses([]release.Status_Code{
			release.Status_DEPLOYED,
			release.Status_FAILED,
			release.Status_DELETING,
			release.Status_PENDING_INSTALL,
			release.Status_PENDING_UPGRADE,
			release.Status_PENDING_ROLLBACK}),
		// helm.ReleaseListLimit(limit),
		// helm.ReleaseListFilter(filter),
		// helm.ReleaseListNamespace(""),
	}

	ops = append(ops, helmapi.ReleaseListFilter(regex))
	listResponse, err := helmClient.ListReleases(ops...)
	if err != nil {
		// if the release is not exist, we should install it.
		return nil, err
	}

	if listResponse == nil {
		glog.Infof("The list response is nil, [%s]", regex)
		return nil, nil
	}

	if listResponse.Releases == nil || len(listResponse.Releases) < 0 {
		glog.Infof("Can not find any release named [%s]", regex)
		return nil, nil
	}

	return listResponse.Releases, nil
}
