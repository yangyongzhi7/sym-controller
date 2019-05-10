/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	migratev1 "k8s.io/sample-controller/pkg/apis/migrate/v1"
)

// FakeMigrates implements MigrateInterface
type FakeMigrates struct {
	Fake *FakeMigrateV1
	ns   string
}

var migratesResource = schema.GroupVersionResource{Group: "migrate.dmall.com", Version: "v1", Resource: "migrates"}

var migratesKind = schema.GroupVersionKind{Group: "migrate.dmall.com", Version: "v1", Kind: "Migrate"}

// Get takes name of the migrate, and returns the corresponding migrate object, and an error if there is any.
func (c *FakeMigrates) Get(name string, options v1.GetOptions) (result *migratev1.Migrate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(migratesResource, c.ns, name), &migratev1.Migrate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*migratev1.Migrate), err
}

// List takes label and field selectors, and returns the list of Migrates that match those selectors.
func (c *FakeMigrates) List(opts v1.ListOptions) (result *migratev1.MigrateList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(migratesResource, migratesKind, c.ns, opts), &migratev1.MigrateList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &migratev1.MigrateList{ListMeta: obj.(*migratev1.MigrateList).ListMeta}
	for _, item := range obj.(*migratev1.MigrateList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested migrates.
func (c *FakeMigrates) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(migratesResource, c.ns, opts))

}

// Create takes the representation of a migrate and creates it.  Returns the server's representation of the migrate, and an error, if there is any.
func (c *FakeMigrates) Create(migrate *migratev1.Migrate) (result *migratev1.Migrate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(migratesResource, c.ns, migrate), &migratev1.Migrate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*migratev1.Migrate), err
}

// Update takes the representation of a migrate and updates it. Returns the server's representation of the migrate, and an error, if there is any.
func (c *FakeMigrates) Update(migrate *migratev1.Migrate) (result *migratev1.Migrate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(migratesResource, c.ns, migrate), &migratev1.Migrate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*migratev1.Migrate), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeMigrates) UpdateStatus(migrate *migratev1.Migrate) (*migratev1.Migrate, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(migratesResource, "status", c.ns, migrate), &migratev1.Migrate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*migratev1.Migrate), err
}

// Delete takes name of the migrate and deletes it. Returns an error if one occurs.
func (c *FakeMigrates) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(migratesResource, c.ns, name), &migratev1.Migrate{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMigrates) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(migratesResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &migratev1.MigrateList{})
	return err
}

// Patch applies the patch and returns the patched migrate.
func (c *FakeMigrates) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *migratev1.Migrate, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(migratesResource, c.ns, name, data, subresources...), &migratev1.Migrate{})

	if obj == nil {
		return nil, err
	}
	return obj.(*migratev1.Migrate), err
}