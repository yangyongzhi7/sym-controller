package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Migrate
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Migrate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MigrateSpec   `json:"spec,omitempty"`
	Status            MigrateStatus `json:"status,omitempty"`
}

// MigrateList
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MigrateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Migrate `json:"items"`
}

// MigrateSpec
type MigrateSpec struct {
	AppName string            `json:"appName,omitempty"`
	Action  MigrateActionType `json:"action,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
	Chart   []byte            `json:"chart,omitempty"`
	// Releases is all of the helm release
	Releases []*ReleasesConfig `json:"releases,omitempty"`
}

type MigrateActionType string

const (
	MigrateActionInstall MigrateActionType = "Install"
	MigrateActionDelete  MigrateActionType = "Delete"
)

// ReleasesConfig
type ReleasesConfig struct {
	// Name is the name of the release
	Name string `json:"name,omitempty"`
	// Namespace is the kubernetes namespace of the release.
	Namespace string `json:"namespace,omitempty"`
	// Config supplies values to the parametrizable templates of a chart.
	Raw    string            `json:"raw,omitempty"`
	Values map[string]string `json:"values,omitempty"`
}

// MigrateStatus
type MigrateStatus struct {
	Conditions     []MigrateCondition `json:"conditions,omitempty"`
	StartTime      *metav1.Time       `json:"startTime,omitempty"`
	LastUpdateTime *metav1.Time       `json:"lastUpdateTime,omitempty"`
}

type MigrateCondition struct {
	Type               string      `json:"type"`
	Status             string      `json:"status"`
	LastProbeTime      metav1.Time `json:"lastProbeTime,omitempty"`
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	Reason             string      `json:"reason,omitempty"`
	Message            string      `json:"message,omitempty"`
}
