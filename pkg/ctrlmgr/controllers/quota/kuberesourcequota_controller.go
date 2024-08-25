/*
Copyright 2024 Kubeworkz Authors

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

package quota

import (
	"context"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	quotav1 "github.com/saashqdev/kubeworkz/pkg/apis/quota/v1"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/ctrlmgr/options"
	"github.com/saashqdev/kubeworkz/pkg/quota"
	"github.com/saashqdev/kubeworkz/pkg/quota/kube"
)

// KubeResourceQuotaReconciler reconciles a KubeResourceQuota object
type KubeResourceQuotaReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	r := &KubeResourceQuotaReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	return r, nil
}

//+kubebuilder:rbac:groups=quota.kubeworkz.io,resources=kuberesourcequota,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=quota.kubeworkz.io,resources=kuberesourcequota/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=quota.kubeworkz.io,resources=kuberesourcequota/finalizers,verbs=update

// Reconcile of kube resource quota only used for initializing status of kube resource quota
func (r *KubeResourceQuotaReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	clog.Info("Reconcile KubeResourceQuota %v", req.Name)

	kubeQuota := &quotav1.KubeResourceQuota{}
	err := r.Get(ctx, req.NamespacedName, kubeQuota)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	quotaOperator := kube.NewQuotaOperator(r.Client, kubeQuota, nil, ctx)

	if kubeQuota.DeletionTimestamp == nil {
		if err := r.ensureFinalizer(ctx, kubeQuota); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if err := r.removeFinalizer(ctx, kubeQuota, quotaOperator); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// init status of kube resource kubeQuota when create
	err = r.initKubeQuotaStatus(ctx, kubeQuota)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.ensureSpecAndStatusConsistent(ctx, kubeQuota)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, quotaOperator.UpdateParentStatus(false)
}

func (r *KubeResourceQuotaReconciler) ensureFinalizer(ctx context.Context, kubeQuota *quotav1.KubeResourceQuota) error {
	if !controllerutil.ContainsFinalizer(kubeQuota, quota.Finalizer) {
		controllerutil.AddFinalizer(kubeQuota, quota.Finalizer)
		if err := r.Update(ctx, kubeQuota); err != nil {
			clog.Warn("add finalizer to KubeResourceQuota %v failed: %v", kubeQuota.Name, err)
			return err
		}
	}
	return nil
}

func (r *KubeResourceQuotaReconciler) removeFinalizer(ctx context.Context, kubeQuota *quotav1.KubeResourceQuota, quotaOperator quota.Interface) error {
	if controllerutil.ContainsFinalizer(kubeQuota, quota.Finalizer) {
		clog.Info("delete KubeResourceQuota %v", kubeQuota.Name)
		err := quotaOperator.UpdateParentStatus(true)
		if err != nil {
			clog.Error("update parent status of KubeResourceQuota %v failed: %v", kubeQuota.Name, err)
			return err
		}
		controllerutil.RemoveFinalizer(kubeQuota, quota.Finalizer)
		err = r.Update(ctx, kubeQuota)
		if err != nil {
			clog.Warn("delete finalizer to KubeResourceQuota %v failed: %v", kubeQuota.Name, err)
			return err
		}
	}
	return nil
}

func (r *KubeResourceQuotaReconciler) initKubeQuotaStatus(ctx context.Context, kubeQuota *quotav1.KubeResourceQuota) error {
	if kubeQuota.Status.Used != nil && kubeQuota.Status.Hard != nil {
		return nil
	}

	kube.InitStatus(kubeQuota)

	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		newQuota := &quotav1.KubeResourceQuota{}
		err := r.Get(ctx, types.NamespacedName{Name: kubeQuota.Name}, newQuota)
		if err != nil {
			return err
		}
		newQuota.Status = kubeQuota.Status
		err = r.Status().Update(ctx, newQuota, &client.SubResourceUpdateOptions{})
		if err != nil {
			return err
		}
		return nil
	})
}

func (r *KubeResourceQuotaReconciler) ensureSpecAndStatusConsistent(ctx context.Context, kubeQuota *quotav1.KubeResourceQuota) error {
	needUpdate := false

	// ensure used field
	used, updateUsed := r.ifUpdateUsed(kubeQuota.Spec.Hard, kubeQuota.Status.Used)
	if updateUsed {
		kubeQuota.Status.Used = used
		needUpdate = true
	}

	// ensure status hard
	if !reflect.DeepEqual(kubeQuota.Spec.Hard, kubeQuota.Status.Hard) {
		kubeQuota.Status.Hard = kubeQuota.Spec.Hard
		needUpdate = true
	}

	if needUpdate {
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			newQuota := &quotav1.KubeResourceQuota{}
			err := r.Get(ctx, types.NamespacedName{Name: kubeQuota.Name}, newQuota)
			if err != nil {
				return err
			}
			newQuota.Status = kubeQuota.Status
			err = r.Status().Update(ctx, newQuota, &client.SubResourceUpdateOptions{})
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// ifUpdateUsed keep resource of hard and used same
func (r *KubeResourceQuotaReconciler) ifUpdateUsed(hard, used v1.ResourceList) (v1.ResourceList, bool) {
	needUpdate := false
	for rsName := range hard {
		if _, ok := used[rsName]; !ok {
			needUpdate = true
			used[rsName] = quota.ZeroQ()
		}
	}
	return used, needUpdate
}

// SetupWithManager sets up the controller with the Manager.
func SetupWithManager(mgr ctrl.Manager, _ *options.Options) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}

	// filter update event
	predicateFunc := predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(updateEvent event.UpdateEvent) bool {
			oldObj, ok := updateEvent.ObjectOld.(*quotav1.KubeResourceQuota)
			if !ok {
				return false
			}
			newObj, ok := updateEvent.ObjectNew.(*quotav1.KubeResourceQuota)
			if !ok {
				return false
			}
			if oldObj.DeletionTimestamp != nil || newObj.DeletionTimestamp != nil {
				return true
			}
			if reflect.DeepEqual(oldObj.Spec, newObj.Spec) {
				return false
			}
			return true
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return true
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return true
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&quotav1.KubeResourceQuota{}).
		WithEventFilter(predicateFunc).
		Complete(r)
}
