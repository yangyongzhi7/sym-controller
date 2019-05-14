package labels

import (
	"fmt"
)

const (
	LabelCreatedBy   = "createdBy"
	LabelClusterName = "clusterName"
	LabelLdcName     = "ldc"
	LabelAzName      = "az"
)

const (
	ControllerName = "sym-controller"
)

func GetCrdLabels() map[string]string {
	return map[string]string{
		LabelCreatedBy: ControllerName,
	}
}

func GetCrdLabelSelector() string {
	return fmt.Sprintf("%v=%v", LabelCreatedBy, ControllerName)
}

func MakeHelmReleaseFilter(appName string) string {
	if appName == "" || appName == "all" {
		return ""
	}
	return fmt.Sprintf("^%s(-gz|-rz).*(-blue|-green)$", appName)
}
