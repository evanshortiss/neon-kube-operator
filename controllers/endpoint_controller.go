/*
Copyright 2023.

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

package controllers

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	neontechv1alpha1 "github.com/evanshortiss/neon-kube-operator/api/v1alpha1"
	"github.com/evanshortiss/neon-kube-operator/neon"
)

// EndpointReconciler reconciles a Endpoint object
type EndpointReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	NeonClient *neon.Client
}

//+kubebuilder:rbac:groups=neon.tech,resources=endpoints,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=neon.tech,resources=endpoints/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=neon.tech,resources=endpoints/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Endpoint object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *EndpointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error

	_ = log.FromContext(ctx)

	e := &neontechv1alpha1.Endpoint{}
	if err = r.Client.Get(ctx, req.NamespacedName, e); err != nil {
		return ctrl.Result{}, err
	}

	if err = AddFinalizer(ctx, r.Client, e); err != nil {
		return ctrl.Result{}, err
	}

	if e.DeletionTimestamp != nil {
		_ = r.updateState(ctx, e, neontechv1alpha1.EndpointStateDeleting)
		if err = r.ExecuteFinalizer(ctx, e); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	err = r.reconcile(ctx, e)

	tries := 5
	for tries < 5 {
		_, updateErr := Update(ctx, r.Client, e, func() error { return nil })
		if updateErr == nil {
			break
		}

		if tries == 4 {
			return ctrl.Result{}, updateErr
		}
	}

	return ctrl.Result{}, err
}

func (r *EndpointReconciler) ExecuteFinalizer(ctx context.Context, endpoint *neontechv1alpha1.Endpoint) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling deletion of endpoint", "name", endpoint.Name)
	if _, err := r.NeonClient.DeleteEndpoint(ctx, endpoint); err != nil {
		return err
	}
	if ok := controllerutil.RemoveFinalizer(endpoint, neonFinalizer); ok {
		if err := r.Update(ctx, endpoint); err != nil {
			return err
		}
		logger.Info("Finalizer removed from endpoint", "name", endpoint.Name)
	}
	return nil
}

func (r *EndpointReconciler) reconcile(ctx context.Context, endpoint *neontechv1alpha1.Endpoint) error {
	logger := log.FromContext(ctx)
	_, err := r.NeonClient.GetEndpoint(ctx, endpoint.Name, &endpoint.Spec)

	shouldCreate := false
	if err != nil {
		if errors.Is(err, neon.BranchNotFound) {
			return err
		}
		shouldCreate = true
	}

	if !shouldCreate {
		endpoint.Status.State = neontechv1alpha1.EndpointStateCreated
		return nil
	}

	logger.Info("Creating endpoint", "name", endpoint.Name)
	_, err = r.NeonClient.CreateEndpoint(ctx, &endpoint.Spec)
	if err != nil {
		return err
	}

	endpoint.Status.State = neontechv1alpha1.EndpointStateCreated

	return nil
}

func (r *EndpointReconciler) updateStatusIfChanged(ctx context.Context, old, new *neontechv1alpha1.Endpoint) error {
	if old.Status.State == new.Status.State {
		return nil
	}
	return r.Client.Status().Update(ctx, new)
}

func (r *EndpointReconciler) updateState(ctx context.Context, endpoint *neontechv1alpha1.Endpoint, state neontechv1alpha1.EndpointState) error {
	endpoint.Status.State = state
	return r.Client.Status().Update(ctx, endpoint)
}

// SetupWithManager sets up the controller with the Manager.
func (r *EndpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&neontechv1alpha1.Endpoint{}).
		Complete(r)
}
