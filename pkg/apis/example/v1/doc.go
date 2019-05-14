// Package v1 is the v1 version of the API.

// +k8s:deepcopy-gen=package,register
// +groupName=example.dmall.com
package v1

const (
	// GroupName is the group name use in this package
	GroupName = "example.dmall.com"
	// ResourceVersion represent the resource version
	ResourceVersion = "v1"
)
