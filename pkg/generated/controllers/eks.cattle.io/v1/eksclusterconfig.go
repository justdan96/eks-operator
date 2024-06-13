/*
Copyright 2019 Wrangler Sample Controller Authors

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

// Code generated by main. DO NOT EDIT.

package v1

import (
	"context"
	"sync"
	"time"

	v1 "github.com/rancher/eks-operator/pkg/apis/eks.cattle.io/v1"
	"github.com/rancher/wrangler/v3/pkg/apply"
	"github.com/rancher/wrangler/v3/pkg/condition"
	"github.com/rancher/wrangler/v3/pkg/generic"
	"github.com/rancher/wrangler/v3/pkg/kv"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// EKSClusterConfigController interface for managing EKSClusterConfig resources.
type EKSClusterConfigController interface {
	generic.ControllerInterface[*v1.EKSClusterConfig, *v1.EKSClusterConfigList]
}

// EKSClusterConfigClient interface for managing EKSClusterConfig resources in Kubernetes.
type EKSClusterConfigClient interface {
	generic.ClientInterface[*v1.EKSClusterConfig, *v1.EKSClusterConfigList]
}

// EKSClusterConfigCache interface for retrieving EKSClusterConfig resources in memory.
type EKSClusterConfigCache interface {
	generic.CacheInterface[*v1.EKSClusterConfig]
}

// EKSClusterConfigStatusHandler is executed for every added or modified EKSClusterConfig. Should return the new status to be updated
type EKSClusterConfigStatusHandler func(obj *v1.EKSClusterConfig, status v1.EKSClusterConfigStatus) (v1.EKSClusterConfigStatus, error)

// EKSClusterConfigGeneratingHandler is the top-level handler that is executed for every EKSClusterConfig event. It extends EKSClusterConfigStatusHandler by a returning a slice of child objects to be passed to apply.Apply
type EKSClusterConfigGeneratingHandler func(obj *v1.EKSClusterConfig, status v1.EKSClusterConfigStatus) ([]runtime.Object, v1.EKSClusterConfigStatus, error)

// RegisterEKSClusterConfigStatusHandler configures a EKSClusterConfigController to execute a EKSClusterConfigStatusHandler for every events observed.
// If a non-empty condition is provided, it will be updated in the status conditions for every handler execution
func RegisterEKSClusterConfigStatusHandler(ctx context.Context, controller EKSClusterConfigController, condition condition.Cond, name string, handler EKSClusterConfigStatusHandler) {
	statusHandler := &eKSClusterConfigStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, generic.FromObjectHandlerToHandler(statusHandler.sync))
}

// RegisterEKSClusterConfigGeneratingHandler configures a EKSClusterConfigController to execute a EKSClusterConfigGeneratingHandler for every events observed, passing the returned objects to the provided apply.Apply.
// If a non-empty condition is provided, it will be updated in the status conditions for every handler execution
func RegisterEKSClusterConfigGeneratingHandler(ctx context.Context, controller EKSClusterConfigController, apply apply.Apply,
	condition condition.Cond, name string, handler EKSClusterConfigGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &eKSClusterConfigGeneratingHandler{
		EKSClusterConfigGeneratingHandler: handler,
		apply:                             apply,
		name:                              name,
		gvk:                               controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterEKSClusterConfigStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type eKSClusterConfigStatusHandler struct {
	client    EKSClusterConfigClient
	condition condition.Cond
	handler   EKSClusterConfigStatusHandler
}

// sync is executed on every resource addition or modification. Executes the configured handlers and sends the updated status to the Kubernetes API
func (a *eKSClusterConfigStatusHandler) sync(key string, obj *v1.EKSClusterConfig) (*v1.EKSClusterConfig, error) {
	if obj == nil {
		return obj, nil
	}

	origStatus := obj.Status.DeepCopy()
	obj = obj.DeepCopy()
	newStatus, err := a.handler(obj, obj.Status)
	if err != nil {
		// Revert to old status on error
		newStatus = *origStatus.DeepCopy()
	}

	if a.condition != "" {
		if errors.IsConflict(err) {
			a.condition.SetError(&newStatus, "", nil)
		} else {
			a.condition.SetError(&newStatus, "", err)
		}
	}
	if !equality.Semantic.DeepEqual(origStatus, &newStatus) {
		if a.condition != "" {
			// Since status has changed, update the lastUpdatedTime
			a.condition.LastUpdated(&newStatus, time.Now().UTC().Format(time.RFC3339))
		}

		var newErr error
		obj.Status = newStatus
		newObj, newErr := a.client.UpdateStatus(obj)
		if err == nil {
			err = newErr
		}
		if newErr == nil {
			obj = newObj
		}
	}
	return obj, err
}

type eKSClusterConfigGeneratingHandler struct {
	EKSClusterConfigGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
	seen  sync.Map
}

// Remove handles the observed deletion of a resource, cascade deleting every associated resource previously applied
func (a *eKSClusterConfigGeneratingHandler) Remove(key string, obj *v1.EKSClusterConfig) (*v1.EKSClusterConfig, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v1.EKSClusterConfig{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	if a.opts.UniqueApplyForResourceVersion {
		a.seen.Delete(key)
	}

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

// Handle executes the configured EKSClusterConfigGeneratingHandler and pass the resulting objects to apply.Apply, finally returning the new status of the resource
func (a *eKSClusterConfigGeneratingHandler) Handle(obj *v1.EKSClusterConfig, status v1.EKSClusterConfigStatus) (v1.EKSClusterConfigStatus, error) {
	if !obj.DeletionTimestamp.IsZero() {
		return status, nil
	}

	objs, newStatus, err := a.EKSClusterConfigGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}
	if !a.isNewResourceVersion(obj) {
		return newStatus, nil
	}

	err = generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
	if err != nil {
		return newStatus, err
	}
	a.storeResourceVersion(obj)
	return newStatus, nil
}

// isNewResourceVersion detects if a specific resource version was already successfully processed.
// Only used if UniqueApplyForResourceVersion is set in generic.GeneratingHandlerOptions
func (a *eKSClusterConfigGeneratingHandler) isNewResourceVersion(obj *v1.EKSClusterConfig) bool {
	if !a.opts.UniqueApplyForResourceVersion {
		return true
	}

	// Apply once per resource version
	key := obj.Namespace + "/" + obj.Name
	previous, ok := a.seen.Load(key)
	return !ok || previous != obj.ResourceVersion
}

// storeResourceVersion keeps track of the latest resource version of an object for which Apply was executed
// Only used if UniqueApplyForResourceVersion is set in generic.GeneratingHandlerOptions
func (a *eKSClusterConfigGeneratingHandler) storeResourceVersion(obj *v1.EKSClusterConfig) {
	if !a.opts.UniqueApplyForResourceVersion {
		return
	}

	key := obj.Namespace + "/" + obj.Name
	a.seen.Store(key, obj.ResourceVersion)
}
