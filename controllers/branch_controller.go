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

const (
	neonFinalizer = "neon.tech/finalizer"
)

// BranchReconciler reconciles a Branch object
type BranchReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	NeonClient *neon.Client
}

//+kubebuilder:rbac:groups=neon.tech,resources=branches,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=neon.tech,resources=branches/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=neon.tech,resources=branches/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Branch object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *BranchReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	b := &neontechv1alpha1.Branch{}
	err := r.Client.Get(ctx, req.NamespacedName, b)
	if err != nil {
		return ctrl.Result{}, err
	}
	if err = AddFinalizer(ctx, r.Client, b); err != nil {
		return ctrl.Result{}, err
	}

	if b.DeletionTimestamp != nil {
		_ = r.updateState(ctx, b, neontechv1alpha1.BranchStateDeleting)
		if err := r.ExecuteFinalizer(ctx, b); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	err = r.reconcile(ctx, b)

	tries := 5
	for tries < 5 {
		_, updateErr := Update(ctx, r.Client, b, func() error { return nil })
		if updateErr == nil {
			break
		}

		if tries == 4 {
			return ctrl.Result{}, updateErr
		}
	}

	return ctrl.Result{}, err
}

func (r *BranchReconciler) ExecuteFinalizer(ctx context.Context, branch *neontechv1alpha1.Branch) error {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling deletion of branch", "name", branch.Name)
	if _, err := r.NeonClient.DeleteBranch(ctx, branch); err != nil {
		return err
	}
	if ok := controllerutil.RemoveFinalizer(branch, neonFinalizer); ok {
		if err := r.Update(ctx, branch); err != nil {
			return err
		}
		logger.Info("Finalizer removed from branch", "name", branch.Name)
	}
	return nil
}

func (r *BranchReconciler) reconcile(ctx context.Context, branch *neontechv1alpha1.Branch) error {
	logger := log.FromContext(ctx)
	resp, err := r.NeonClient.GetBranch(ctx, branch.Name, &branch.Spec)
	shouldCreate := false
	if err != nil {
		if errors.Is(err, neon.BranchNotFound) {
			return err
		}

		shouldCreate = true
	}
	if !shouldCreate {
		branch.Status = neon.NewBranchStatus(resp)
		return nil
	}
	logger.Info("Creating branch", "name", branch.Name)
	resp, err = r.NeonClient.CreateBranch(ctx, &branch.Spec)
	if err != nil {
		return err
	}
	branch.Status = neon.NewBranchStatus(resp)
	branch.Status.State = neontechv1alpha1.BranchStateCreated
	return nil
}

func AddFinalizer(ctx context.Context, c client.Client, object client.Object) error {
	logger := log.FromContext(ctx)
	if !controllerutil.ContainsFinalizer(object, neonFinalizer) && object.GetDeletionTimestamp() == nil {
		controllerutil.AddFinalizer(object, neonFinalizer)
		err := c.Update(ctx, object)
		if err != nil {
			return err
		}
		logger.Info("Finalizer added into custom resource successfully")
	}
	return nil
}

func (r *BranchReconciler) updateStatusIfChanged(ctx context.Context, old, new *neontechv1alpha1.Branch) error {
	if old.Status.State == new.Status.State {
		return nil
	}
	return r.Client.Status().Update(ctx, new)
}

func (r *BranchReconciler) updateState(ctx context.Context, branch *neontechv1alpha1.Branch, state neontechv1alpha1.BranchState) error {
	branch.Status.State = state
	return r.Client.Status().Update(ctx, branch)
}

// SetupWithManager sets up the controller with the Manager.
func (r *BranchReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&neontechv1alpha1.Branch{}).
		Complete(r)
}
