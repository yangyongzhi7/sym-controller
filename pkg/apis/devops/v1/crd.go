package v1

import (
	"github.com/yangyongzhi/sym-operator/pkg/labels"
	crdapi "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const MigrateResourceName = ResourcePlural + "." + GroupName

// NewCrd defines a Resource as a k8s CR
func NewCrdMigrate() *crdapi.CustomResourceDefinition {
	crd := &crdapi.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name:   MigrateResourceName,
			Labels: labels.GetCrdLabels(),
		},
		Spec: crdapi.CustomResourceDefinitionSpec{
			Group:   GroupName,
			Version: SchemeGroupVersion.Version,
			Scope:   crdapi.NamespaceScoped,
			Names: crdapi.CustomResourceDefinitionNames{
				Plural:     ResourcePlural,
				Singular:   ResourceSingular,
				Kind:       ResourceKind,
				ShortNames: []string{"mgte"},
			},
		},
	}

	return crd
}
