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

package controllers

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	v1 "github.com/saashqdev/kubeworkz/pkg/apis/quota/v1"
	tenantv1 "github.com/saashqdev/kubeworkz/pkg/apis/tenant/v1"
	"github.com/saashqdev/kubeworkz/pkg/clog"
	"github.com/saashqdev/kubeworkz/pkg/utils/constants"
	"github.com/saashqdev/kubeworkz/pkg/utils/env"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ reconcile.Reconciler = &TenantReconciler{}

const (
	// Default timeouts to be used in TimeoutContext
	waitInterval = 2 * time.Second
	waitTimeout  = 120 * time.Second
)

// TenantReconciler reconciles a Tenant object
type TenantReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func newReconciler(mgr manager.Manager) (*TenantReconciler, error) {
	r := &TenantReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}
	return r, nil
}

//+kubebuilder:rbac:groups=tenant.kubeworkz.io,resources=tenants,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tenant.kubeworkz.io,resources=tenants/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tenant.kubeworkz.io,resources=tenants/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the Tenant object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *TenantReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := clog.WithName("reconcile").WithValues("tenant", req.NamespacedName)

	// get tenant info
	tenant := tenantv1.Tenant{}
	err := r.Client.Get(ctx, req.NamespacedName, &tenant)
	if err != nil {
		if errors.IsNotFound(err) {
			return r.deleteTenant(req.Name)
		}
		log.Warn("get tenant fail, %v", err)
		return ctrl.Result{}, nil
	}

	needUpdate := false

	// if .spec.namespace not equal the standard name
	nsName := constants.TenantNsPrefix + req.Name
	if tenant.Spec.Namespace != nsName {
		tenant.Spec.Namespace = nsName
		needUpdate = true
	}

	// if annotation not content kubeworkz.io/sync, add it
	ano := tenant.Annotations
	if ano == nil {
		ano = make(map[string]string)
	}
	if _, ok := ano[constants.SyncAnnotation]; !ok {
		ano[constants.SyncAnnotation] = "1"
		tenant.Annotations = ano
		needUpdate = true
	}

	if needUpdate {
		err = r.Client.Update(ctx, &tenant)
		if err != nil {
			log.Error("update tenant fail, %v", err)
			return ctrl.Result{}, err
		}
	}

	if env.CreateHNCNs() {
		err = r.crateTenantNamespace(ctx, tenant.Name)
		if err != nil {
			clog.Error(err.Error())
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}
func (r *TenantReconciler) deleteTenant(tenantName string) (ctrl.Result, error) {
	// get projects in tenant
	// delete namespace of tenant
	if err := r.deleteNSofTenant(tenantName); err != nil {
		return ctrl.Result{}, err
	}
	// delete kubeResourceQuota of tenant
	err := r.deleteKubeResourceQuotaOfTenant(tenantName)
	return ctrl.Result{}, err
}

func (r *TenantReconciler) deleteNSofTenant(tenantName string) error {
	namespace := &corev1.Namespace{}
	name := constants.TenantNsPrefix + tenantName
	ctx := context.Background()
	if err := r.Client.Get(ctx, types.NamespacedName{Name: name}, namespace); err != nil {
		if errors.IsNotFound(err) {
			return nil
		} else {
			clog.Error("get namespace of tenant err: %s", err.Error())
			return fmt.Errorf("get namespace of tenant err")
		}
	}
	if err := r.Client.Delete(ctx, namespace); err != nil {
		clog.Error("delete namespace of tenant err: %s", err.Error())
		return err
	}
	err := wait.PollUntilContextTimeout(ctx, waitInterval, waitTimeout, false,
		func(ctx context.Context) (bool, error) {
			e := r.Client.Get(ctx, types.NamespacedName{Name: name}, namespace)
			if errors.IsNotFound(e) {
				return true, nil
			} else {
				return false, nil
			}
		})
	if err != nil {
		clog.Error("wait for delete namespace of tenant err: %s", err.Error())
		return err
	}
	return nil
}

func (r *TenantReconciler) deleteKubeResourceQuotaOfTenant(tenantName string) error {
	quota := v1.KubeResourceQuota{}
	err := r.Client.DeleteAllOf(context.TODO(), &quota, client.MatchingLabels{constants.TenantLabel: tenantName})
	if err != nil && !errors.IsNotFound(err) {
		clog.Error("delete kube resource quota error， tenant name: %s, error: %s", tenantName, err.Error())
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func SetupWithManager(mgr ctrl.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenantv1.Tenant{}).
		Complete(r)
}

func (r *TenantReconciler) crateTenantNamespace(ctx context.Context, tenant string) error {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("kubeworkz-tenant-%v", tenant),
			Annotations: map[string]string{"hnc.x-k8s.io/ns": "true"},
			Labels: map[string]string{
				constants.HncIncludedNsLabel:                                       "true",
				fmt.Sprintf("kubeworkz-tenant-%v.tree.hnc.x-k8s.io/depth", tenant): "0",
			},
		},
	}

	err := r.Create(ctx, ns)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	return nil
}
