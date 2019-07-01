/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"fmt"
	"github.com/yangyongzhi/sym-operator/pkg/apis/devops/v1"
	"github.com/yangyongzhi/sym-operator/pkg/constant"
	"github.com/yangyongzhi/sym-operator/pkg/helm"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog"
	"strings"
	"time"

	samplev1alpha1 "github.com/yangyongzhi/sym-operator/pkg/apis/example/v1"
	clientset "github.com/yangyongzhi/sym-operator/pkg/client/clientset/versioned"
	samplescheme "github.com/yangyongzhi/sym-operator/pkg/client/clientset/versioned/scheme"
	informers "github.com/yangyongzhi/sym-operator/pkg/client/informers/externalversions/devops/v1"
	listers "github.com/yangyongzhi/sym-operator/pkg/client/listers/devops/v1"
)

const controllerAgentName = "symphony-operator"
const migrateNamespace = "default"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced          = "Synced"
	SuccessInstalledStatus = "SuccessInstalledStatus"
	SuccessUpdatedStatus   = "SuccessUpdatedStatus"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ResourceExists    = "ResourceExists"
	ErrGetRelease     = "ErrGetRelease"
	ErrReleaseContent = "ErrReleaseContent"
	FailInstall       = "FailInstall"
	FailUpdate        = "FailUpdate"
	ErrDeleteRelease  = "ErrDeleteRelease"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Migrate synced successfully"
)

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	// but symclientset is a clientset for our own API group
	kubeclientset kubernetes.Interface
	symclientset  clientset.Interface

	helmClient *helm.Client

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced
	symLister         listers.MigrateLister
	symSynced         cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	symclientset clientset.Interface, helmClient *helm.Client,
	deploymentInformer appsinformers.DeploymentInformer,
	symInformer informers.MigrateInformer) *Controller {

	// Create event broadcaster
	// Add sym-migrate-controller types to the default Kubernetes Scheme so Events can be
	// logged for sample-controller types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:     kubeclientset,
		symclientset:      symclientset,
		helmClient:        helmClient,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		symLister:         symInformer.Lister(),
		symSynced:         symInformer.Informer().HasSynced,
		workqueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Sym"),
		recorder:          recorder,
	}

	klog.Info("Setting up event handlers")

	// Set up an event handler for when Foo resources change
	symInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.enqueueMigrate,
		UpdateFunc: func(old, new interface{}) {
			controller.enqueueMigrate(new)
		},
	})

	// Set up an event handler for when Deployment resources change. This
	// handler will lookup the owner of the given Deployment, and if it is
	// owned by a Foo resource will enqueue that Foo resource for
	// processing. This way, we don't need to implement custom logic for
	// handling Deployment resources. More info on this pattern:
	// https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
	deploymentInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: controller.handleObject,
		UpdateFunc: func(old, new interface{}) {
			newDepl := new.(*appsv1.Deployment)
			oldDepl := old.(*appsv1.Deployment)
			if newDepl.ResourceVersion == oldDepl.ResourceVersion {
				// Periodic resync will send update events for all known Deployments.
				// Two different versions of the same Deployment will always have different RVs.
				return
			}
			controller.handleObject(new)
		},
		DeleteFunc: controller.handleObject,
	})

	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Symphony operator...")
	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync...")
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.symSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)
		var key string
		var ok bool
		// We expect strings to come off the workqueue. These are of the
		// form namespace/name. We do this as the delayed nature of the
		// workqueue means the items in the informer cache may actually be
		// more up to date that when the item was initially put onto the
		// workqueue.
		if key, ok = obj.(string); !ok {
			// As the item in the workqueue is actually invalid, we call
			// Forget here else we'd go into a loop of attempting to
			// process a work item that is invalid.
			c.workqueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
			return nil
		}
		// Run the syncHandler, passing it the namespace/name string of the
		// Foo resource to be synced.
		if err := c.syncHandler(key); err != nil {
			// Put the item back on the workqueue to handle any transient errors.
			c.workqueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", key, err.Error())
		}
		// Finally, if no error occurs we Forget this item so it does not
		// get queued again until another change happens.
		c.workqueue.Forget(obj)
		klog.Infof("Successfully synced '%s'", key)
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key string) error {
	klog.Infof("Start sync handler method, key : '%s'", key)

	// Convert the namespace/name string into a distinct namespace and name
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	// Get the Foo resource with this namespace/name
	migrate, err := c.symLister.Migrates(namespace).Get(name)
	if err != nil {
		// The Foo resource may no longer exist, in which case we stop
		// processing.
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("migrate '%s' in work queue no longer exists", key))
			return nil
		}

		return err
	}

	if migrate == nil {
		klog.Infof("Can not find the migrate with key: '%s'", key)
		return nil
	}
	if migrate.DeletionTimestamp != nil {
		klog.Infof("The migrate has been deleted, key: '%s'", key)
		return nil
	}

	appName := migrate.Spec.AppName
	klog.Infof("The appName of this migrate : '%s'", appName)
	if appName == "" {
		// We choose to absorb the error here as the worker would requeue the
		// resource otherwise. Instead, the next time the resource is updated
		// the resource will be queued again.
		utilruntime.HandleError(fmt.Errorf("%s: app name must be specified", key))
		return nil
	}

	// Find all deployment with the app name, always there should be two deployments with this app name (blue & green).
	//deployment, err := c.deploymentsLister.Deployments(object.GetNamespace()).Get(object.GetName())
	//r, _ := labels.NewRequirement("app", selection.Equals, []string{appName})
	labelSet := labels.Set{}
	labelSet[constant.AppLabel] = appName
	deployments, err := c.deploymentsLister.Deployments(migrate.GetNamespace()).List(labels.SelectorFromSet(labelSet))
	// If the resource doesn't exist, we'll create it
	//if errors.IsNotFound(err) {
	//	deployment, err = c.kubeclientset.AppsV1().Deployments(foo.Namespace).Create(newDeployment(foo))
	//}

	// If an error occurs during Get/Create, we'll requeue the item so we can
	// attempt processing again later. This could have been caused by a
	// temporary network failure, or any other transient reason.
	if err != nil {
		klog.Infof("Can not find the deployments for migrate [%s], ignore it.", migrate.GetName())
		return nil
	}
	action := migrate.Spec.Action
	klog.Infof("The action of migrate [%s]: '%s'", migrate.GetName(), action)

	if action == v1.MigrateActionInstall {
		c.installReleases(migrate)
		c.syncInstallMigrateStatus(migrate, deployments)
	} else if action == v1.MigrateActionDelete {
		c.deleteReleases(migrate)
		c.syncDeleteMigrateStatus(migrate, deployments)
	} else if action == v1.MigrateActionUpdate {
		revisions := c.updateReleases(migrate, deployments)
		c.syncUpdateMigrateStatus(migrate, deployments, revisions)
	}

	c.recorder.Event(migrate, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	//+", "+strconv.Itoa(rand.Int())
	return nil
}

// Install release
func (c *Controller) installReleases(migrate *v1.Migrate) {
	if migrate.Status.Finished == constant.ConditionStatusTrue {
		klog.Infof("Migrate has been execute successfully: '%s'", migrate.Name)
		return
	}

	releases := migrate.Spec.Releases
	klog.Infof("The raw data as base64 format of this migrate : '%s'", releases)

	for _, release := range releases {
		rlsName := release.Name
		rlsNamespace := release.Namespace
		klog.Infof("Ready to install release [%s] in namespace [%s]", rlsName, rlsNamespace)
		getRelease, err := c.helmClient.GetRelease(rlsName)
		if err != nil {
			//status, _ := status.FromError(err)
			c.recorder.Event(migrate, corev1.EventTypeWarning, ErrGetRelease,
				fmt.Sprintf("Error - Get the release [%s] info : %s", rlsName, err))
			continue
		}

		// you should install it.
		if getRelease == nil {
			//releaseResponse.GetRelease().Chart
			installResponse, err := c.helmClient.InstallRelease(rlsNamespace, rlsName, migrate.Spec.Chart, release.Raw)
			if err != nil {
				c.recorder.Event(migrate, corev1.EventTypeWarning, ErrReleaseContent,
					fmt.Sprintf("Install release [%s] has an error : %s", rlsName, err))
			} else {
				c.recorder.Event(migrate, corev1.EventTypeNormal, SuccessInstalledStatus,
					fmt.Sprintf("Install release [%s] successfully, version : %d", rlsName, installResponse.Release.Version))
			}
		} else {
			// the release has been exist, you can only update it.
			c.recorder.Event(migrate, corev1.EventTypeNormal, ResourceExists,
				fmt.Sprintf("Release [%s] has been installed, operater don't need to do anything.", rlsName))
		}
	}
}

// Update release
func (c *Controller) updateReleases(migrate *v1.Migrate, deployments []*appsv1.Deployment) map[string]int32 {
	klog.Infof("##### Start to update migrate: '%s'", migrate.Name)
	if migrate.Status.Finished == constant.ConditionStatusTrue {
		klog.Infof("##### The status of migrate has been set as true: '%s', so no need to do anything.", migrate.Name)
		return nil
	}

	revisions := map[string]int32{}
	releases := migrate.Spec.Releases
	releaseIterator := [2]string{constant.BlueGroup, constant.GreenGroup}
	for _, releaseGroup := range releaseIterator {
		isNeed := false
		var currentRelease *v1.ReleasesConfig
		for _, release := range releases {
			if strings.Contains(release.Name, "-"+releaseGroup) {
				isNeed = true
				currentRelease = release
			}
		}

		// This release which belongs to this group must be exist.
		if isNeed {
			rlsName := currentRelease.Name
			rlsNamespace := currentRelease.Namespace
			klog.Infof("##### Release [%s]-[%s] has been defined in Spec, so we should update it", rlsName, rlsNamespace)
			getRelease, err := c.helmClient.GetRelease(rlsName)
			if err != nil {
				//status, _ := status.FromError(err)
				c.recorder.Event(migrate, corev1.EventTypeWarning, ErrGetRelease,
					fmt.Sprintf("Error - find the release [%s] info : %s, so can not update", rlsName, err))
				continue
			}

			// if this release has been exist, we should update it.
			if getRelease != nil {
				klog.Infof("##### We already found the release [%s], then we will update it immediately.", rlsName)
				if migrate.Status.ReleaseRevision[rlsName] == getRelease.Version {
					klog.Infof("##### Release version in the status [%s] has been updated to version [%d], no need to do anything.",
						rlsName, migrate.Status.ReleaseRevision[rlsName])
					revisions[rlsName] = getRelease.Version
					continue
				}

				// The version is not same as the one has been aved in status.
				updateResponse, err := c.helmClient.UpdateRelease(rlsName, migrate.Spec.Chart, currentRelease.Raw)
				if err != nil {
					c.recorder.Event(migrate, corev1.EventTypeWarning, ErrReleaseContent,
						fmt.Sprintf("Update release [%s] has an error : %s", rlsName, err))
				} else {
					revisions[rlsName] = updateResponse.Release.Version
					c.recorder.Event(migrate, corev1.EventTypeNormal, SuccessUpdatedStatus,
						fmt.Sprintf("Update release [%s] successfully, version : %d", rlsName, updateResponse.Release.Version))
				}
				continue
			} else {
				klog.Infof("##### Can not find release [%s] when you want to update it, so you should use installing instead of updating.", rlsName)
				installResponse, err := c.helmClient.InstallRelease(rlsNamespace, rlsName, migrate.Spec.Chart, currentRelease.Raw)
				if err != nil {
					c.recorder.Event(migrate, corev1.EventTypeWarning, ErrReleaseContent,
						fmt.Sprintf("Install release [%s] has an error : %s", rlsName, err))
				} else {
					revisions[rlsName] = installResponse.Release.Version
					c.recorder.Event(migrate, corev1.EventTypeNormal, SuccessInstalledStatus,
						fmt.Sprintf("Install release [%s] successfully, version : %d", rlsName, installResponse.Release.Version))
				}
				continue
			}
		} else {
			klog.Infof("##### We found a release which belong to group [%s] is not in the Spec definition, we should delete it.", releaseGroup)
			// if this release which belongs to current group is not be defined, we must delete it.
			releases, err := c.helmClient.FilterReleases(fmt.Sprintf("^%s(-gz|-rz).*(-%s)$", migrate.Spec.AppName, releaseGroup))
			if err != nil {
				// if the release is not exist, we should install it.
				c.recorder.Event(migrate, corev1.EventTypeWarning, ErrGetRelease,
					fmt.Sprintf("Error - list the release info : %s, so can not update", err.Error()))
				continue
			}

			if releases == nil {
				klog.Infof("##### Found no release for group [%s] which will be deleted, no need to anything.", releaseGroup)
				continue
			}

			if len(releases) == 1 {
				klog.Infof("##### Found the release [%s] which will be deleted", releases[0].Name)
				_, err := c.helmClient.UninstallRelease(releases[0].Name)
				if err != nil {
					c.recorder.Event(migrate, corev1.EventTypeWarning, ErrDeleteRelease,
						fmt.Sprintf("Delete release [%s] has an error : %s", releases[0].Name, err.Error()))
					continue
				} else {
					// Don't save the version of the deleted release into the status.
					c.recorder.Event(migrate, corev1.EventTypeNormal, SuccessSynced,
						fmt.Sprintf("Release [%s] has been deleted successfully.", releases[0].Name))
				}
			} else {
				klog.Infof("##### Found more than one release [%d] which will be deleted, we are not sure which one should be deleted?", len(releases))
			}
		}
	}
	return revisions
}

// Delete releases
func (c *Controller) deleteReleases(migrate *v1.Migrate) {
	if migrate.Status.Finished == constant.ConditionStatusTrue {
		klog.Infof("Migrate has been execute successfully: '%s'", migrate.Name)
		return
	}

	releases, err := c.helmClient.FilterReleases(fmt.Sprintf("^%s(-gz|-rz).*(-%s|-%s)$", migrate.Spec.AppName, constant.BlueGroup, constant.GreenGroup))
	if err != nil {
		c.recorder.Event(migrate, corev1.EventTypeWarning, ErrDeleteRelease,
			fmt.Sprintf("Can not find any releases when you want to delete them [%s], error : %s", migrate.Name, err.Error()))
	}

	for _, release := range releases {
		rlsName := release.Name
		rlsNamespace := release.Namespace
		klog.Infof("Ready to delete release [%s] in namespace [%s]", rlsName, rlsNamespace)
		_, err := c.helmClient.UninstallRelease(rlsName)
		if err != nil {
			c.recorder.Event(migrate, corev1.EventTypeWarning, ErrDeleteRelease,
				fmt.Sprintf("Delete release [%s] has an error : %s", rlsName, err.Error()))
			continue
		} else {
			c.recorder.Event(migrate, corev1.EventTypeNormal, SuccessSynced,
				fmt.Sprintf("Release [%s] has been deleted successfully.", rlsName))
		}
	}
}

// enqueueFoo takes a Foo resource and converts it into a namespace/name
// string which is then put onto the work queue. This method should *not* be
// passed resources of any type other than Foo.
func (c *Controller) enqueueMigrate(obj interface{}) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(err)
		return
	}
	c.workqueue.Add(key)
}

// handleObject will take any resource implementing metav1.Object and attempt
// to find the Foo resource that 'owns' it. It does this by looking at the
// objects metadata.ownerReferences field for an appropriate OwnerReference.
// It then enqueues that Foo resource to be processed. If the object does not
// have an appropriate OwnerReference, it will simply be skipped.
func (c *Controller) handleObject(obj interface{}) {
	var object metav1.Object
	var ok bool
	if object, ok = obj.(metav1.Object); !ok {
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object, invalid type"))
			return
		}
		object, ok = tombstone.Obj.(metav1.Object)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("error decoding object tombstone, invalid type"))
			return
		}
		klog.Infof("Recovered deleted object '%s' from tombstone", object.GetName())
	}

	appName := object.GetLabels()[constant.AppLabel]
	klog.Infof("Processing deployment: %s, app name: %s", object.GetName(), appName)

	// Find the migrate with the app name of this deployment in this namespace.
	migrate, err := c.symLister.Migrates(object.GetNamespace()).Get(appName)
	if err != nil {
		klog.Infof("Find migrate task for deployment [%s] has an error: %s", object.GetName(), err.Error())
		return
	}
	if migrate == nil {
		klog.Infof("Can not find a migrate task for deployment: %s, ignore it.", object.GetName())
		return
	}
	if migrate.DeletionTimestamp != nil {
		klog.Infof("Perhaps the migrate which related with this deployment [%s] has been deleted, ignore it.", object.GetName())
		return
	}

	klog.Infof("##### Enqueue ##### enqueue a migrate [%s] due to a event of the deployment [%s], app name: %s", migrate.Name, object.GetName(), appName)
	c.enqueueMigrate(migrate)
	return

}

// Updating the status of a migrate which has been set as a deleting one.
func (c *Controller) syncDeleteMigrateStatus(migrate *v1.Migrate, deployments []*appsv1.Deployment) error {
	migrateCopy := migrate.DeepCopy()
	initialFinished := migrateCopy.Status.Finished
	now := metav1.Now()

	upsertCondition(migrateCopy, v1.MigrateCondition{
		constant.ConcatConditionType(constant.BlueGroup), constant.ConditionStatusTrue,
		now, now, "", "The release has been deleted"})
	upsertCondition(migrateCopy, v1.MigrateCondition{
		constant.ConcatConditionType(constant.GreenGroup), constant.ConditionStatusTrue,
		now, now, "", "The release has been deleted"})

	if deployments != nil && len(deployments) > 0 {
		klog.Infof("===== Wait for the deployments which related to migrate [%s] to disappear.", migrate.GetName())
		for _, deploy := range deployments {
			message := fmt.Sprintf("Check the deployment [%s]: replica:%d, available:%d",
				deploy.GetName(), deploy.Status.Replicas, deploy.Status.AvailableReplicas)
			klog.Info("=====" + message)
			upsertCondition(migrateCopy, v1.MigrateCondition{
				constant.ConcatConditionType(deploy.Labels[constant.GroupLabel]), constant.ConditionStatusFalse,
				now, now, "", "The deployment still exists"})
		}
	}

	calFinalStatus(migrateCopy)
	if initialFinished == constant.ConditionStatusFalse || migrateCopy.Status.Finished == constant.ConditionStatusFalse {
		migrateCopy.Status.LastUpdateTime = &now
	}

	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance

	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.symclientset.DevopsV1().Migrates(migrate.Namespace).Update(migrateCopy)
	//_, err := c.symclientset.DevopsV1().Migrates(migrate.Namespace).UpdateStatus(migrateCopy)
	return err
}

// Updating the status of migrate which has been set as a installing one
func (c *Controller) syncInstallMigrateStatus(migrate *v1.Migrate, deployments []*appsv1.Deployment) error {
	migrateCopy := migrate.DeepCopy()
	initialFinished := migrateCopy.Status.Finished
	now := metav1.Now()

	if deployments != nil && len(deployments) > 0 {
		klog.Infof("===== The deployments belongs to migrate [%s] is not null or not empty, update the status of the INSTALL migrate.", migrate.GetName())
		for _, deploy := range deployments {
			conditionType := constant.ConcatConditionType(deploy.Labels[constant.GroupLabel])
			message := fmt.Sprintf("===== Deployment %s status: replica:%d, available:%d",
				deploy.GetName(), deploy.Status.Replicas, deploy.Status.AvailableReplicas)
			upsertCondition(migrateCopy, v1.MigrateCondition{
				conditionType, constant.ConditionStatusFalse, now, now, "", message})
			klog.Info(message)
			if deploy.Status.Replicas == deploy.Status.AvailableReplicas {
				upsertCondition(migrateCopy, v1.MigrateCondition{
					conditionType, constant.ConditionStatusTrue, now, now, "", message})
			}
		}
	}

	calFinalStatus(migrateCopy)
	if initialFinished == constant.ConditionStatusFalse || migrateCopy.Status.Finished == constant.ConditionStatusFalse {
		migrateCopy.Status.LastUpdateTime = &now
	}
	// Only for test
	//clearConditions(migrateCopy)

	// NEVER modify objects from the store. It's a read-only, local cache.
	// You can use DeepCopy() to make a deep copy of original object and modify this copy
	// Or create a copy manually for better performance

	// If the CustomResourceSubresources feature gate is not enabled,
	// we must use Update instead of UpdateStatus to update the Status block of the Foo resource.
	// UpdateStatus will not allow changes to the Spec of the resource,
	// which is ideal for ensuring nothing other than resource status has been updated.
	_, err := c.symclientset.DevopsV1().Migrates(migrate.Namespace).Update(migrateCopy)
	//_, err := c.symclientset.DevopsV1().Migrates(migrate.Namespace).UpdateStatus(migrateCopy)
	return err
}

// Updating the status of migrate which has been set as a installing one
func (c *Controller) syncUpdateMigrateStatus(migrate *v1.Migrate, deployments []*appsv1.Deployment, revisions map[string]int32) error {
	migrateCopy := migrate.DeepCopy()
	initialFinished := migrateCopy.Status.Finished
	now := metav1.Now()

	if deployments != nil && len(deployments) > 0 {
		klog.Infof("===== The deployments for migrate [%s] is not null or not empty, then update the status of the UPDATE migrate.", migrate.GetName())
		for _, deploy := range deployments {
			var message = ""
			conditionType := constant.ConcatConditionType(deploy.Labels[constant.GroupLabel])

			rlsName := deploy.Spec.Template.Labels[constant.ReleaseLabel]
			var currentRelease *v1.ReleasesConfig
			for _, rls := range migrate.Spec.Releases {
				if rls.Name == rlsName {
					currentRelease = rls
				}
			}
			if currentRelease == nil {
				klog.Infof("===== Can not find release in Spec part with deployment's label %s, so wait for it to disappear.", rlsName)
				upsertCondition(migrateCopy, v1.MigrateCondition{
					conditionType, constant.ConditionStatusFalse, now, now, "",
					fmt.Sprintf("Can not find release in Spec part with deployment's label %s, so wait for it to disappear.", rlsName)})
				continue
			}

			message = fmt.Sprintf("Deployment [%s]'s status: desired replica:%d, available:%d, Migrate replica count:%d",
				deploy.GetName(), deploy.Status.Replicas, deploy.Status.AvailableReplicas, currentRelease.Replicas)
			upsertCondition(migrateCopy, v1.MigrateCondition{
				conditionType, constant.ConditionStatusFalse, now, now, "", message})
			klog.Info("===== " + message)
			if deploy.Status.Replicas == deploy.Status.AvailableReplicas && deploy.Status.AvailableReplicas == currentRelease.Replicas {
				getRelease, err := c.helmClient.GetRelease(rlsName)
				if err != nil {
					klog.Infof("Find release [%s] has an error : %s", rlsName, err.Error())
				} else {
					if migrateCopy.Status.ReleaseRevision == nil {
						message = fmt.Sprintf("The revision information in Status is null, maybe you don't update the release yet. migrate [%s]",
							migrateCopy.Name)
						klog.Info("===== " + message)
						upsertCondition(migrateCopy, v1.MigrateCondition{
							conditionType, constant.ConditionStatusFalse, now, now, "", message})
						continue
					}

					if getRelease != nil && getRelease.Version == migrateCopy.Status.ReleaseRevision[rlsName] {
						upsertCondition(migrateCopy,
							v1.MigrateCondition{conditionType, constant.ConditionStatusTrue, now, now, "", message})
					} else {
						message = fmt.Sprintf("The revision information  [%d] in Status is not equals to the revision  [%d] in tiller, wait for the next updating.",
							migrateCopy.Name)
						klog.Info("===== " + message)
						upsertCondition(migrateCopy, v1.MigrateCondition{
							conditionType, constant.ConditionStatusFalse, now, now, "", message})
					}
				}
			} else {
				message = fmt.Sprintf("Waiting for the deployment [%s] is available if you want to update the Status of migrate [%s]",
					migrateCopy.Name, deploy.Name)
				klog.Info("===== " + message)
			}
		}
	}

	calFinalStatus(migrateCopy)
	if initialFinished == constant.ConditionStatusFalse || migrateCopy.Status.Finished == constant.ConditionStatusFalse {
		migrateCopy.Status.LastUpdateTime = &now
	}

	if revisions != nil && len(revisions) > 0 {
		migrateCopy.Status.ReleaseRevision = revisions
	}

	_, err := c.symclientset.DevopsV1().Migrates(migrate.Namespace).Update(migrateCopy)

	return err
}

// Updating or inserting a condition for this migrate
func upsertCondition(migrateCopy *v1.Migrate, condition v1.MigrateCondition) {
	if len(migrateCopy.Status.Conditions) <= 0 {
		migrateCopy.Status.Conditions = append(migrateCopy.Status.Conditions, condition)
		return
	}

	for i, _ := range migrateCopy.Status.Conditions {
		if migrateCopy.Status.Conditions[i].Type == condition.Type {
			migrateCopy.Status.Conditions[i] = condition
			//c.LastProbeTime = condition.LastProbeTime
			//c.LastTransitionTime = condition.LastTransitionTime
			//c.Message = condition.Message
			//c.Reason = condition.Reason
			return
		}
	}

	migrateCopy.Status.Conditions = append(migrateCopy.Status.Conditions, condition)
}

// You should calculate the final status for this migrate after inserting (update) its conditions.
func calFinalStatus(migrateCopy *v1.Migrate) {
	if len(migrateCopy.Status.Conditions) != len(migrateCopy.Spec.Releases) {
		migrateCopy.Status.Finished = constant.ConditionStatusFalse
		return
	}

	for _, c := range migrateCopy.Status.Conditions {
		if c.Status == constant.ConditionStatusFalse {
			migrateCopy.Status.Finished = constant.ConditionStatusFalse
			return
		}
	}

	migrateCopy.Status.Finished = constant.ConditionStatusTrue
}

//It is always used for a test case if you want to delete all the redundant conditions.
func clearConditions(migrateCopy *v1.Migrate) {
	if len(migrateCopy.Status.Conditions) <= 0 {
		return
	}

	migrateCopy.Status.Conditions = make([]v1.MigrateCondition, 0)

}

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func newDeployment(foo *samplev1alpha1.Foo) *appsv1.Deployment {
	labels := map[string]string{
		"app":        "nginx",
		"controller": foo.Name,
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      foo.Spec.DeploymentName,
			Namespace: foo.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(foo, samplev1alpha1.SchemeGroupVersion.WithKind("Foo")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: foo.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
}
