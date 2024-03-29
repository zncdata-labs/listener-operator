package listenercsi

import (
	"context"
	"time"

	listenersv1alpha1 "github.com/zncdata-labs/listener-operator/api/v1alpha1"
	util "github.com/zncdata-labs/listener-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RBAC struct {
	client client.Client

	cr *listenersv1alpha1.ListenerCSI
}

func NewRBAC(client client.Client, cr *listenersv1alpha1.ListenerCSI) *RBAC {
	return &RBAC{
		client: client,
		cr:     cr,
	}
}

func (r *RBAC) Reconcile(ctx context.Context) (ctrl.Result, error) {

	return r.apply(ctx)
}

func (r *RBAC) apply(ctx context.Context) (ctrl.Result, error) {

	sa, clusterRole, clusterRoleBinding := r.build()

	if mutant, err := util.CreateOrUpdate(ctx, r.client, clusterRole); err != nil {
		return ctrl.Result{}, err
	} else if mutant {
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	if mutant, err := util.CreateOrUpdate(ctx, r.client, sa); err != nil {
		return ctrl.Result{}, err
	} else if mutant {
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	if mutant, err := util.CreateOrUpdate(ctx, r.client, clusterRoleBinding); err != nil {
		return ctrl.Result{}, err
	} else if mutant {
		return ctrl.Result{RequeueAfter: time.Second}, nil
	}

	return ctrl.Result{}, nil

}

func (r *RBAC) build() (*corev1.ServiceAccount, *rbacv1.ClusterRole, *rbacv1.ClusterRoleBinding) {

	sa := r.buildServiceAccount()
	clusterRole := r.buildClusterRole()
	clusterRoleBinding := r.buildClusterRoleBinding()

	return sa, clusterRole, clusterRoleBinding
}

func (r *RBAC) buildServiceAccount() *corev1.ServiceAccount {

	obj := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CSIServiceAccountName,
			Namespace: r.cr.GetNamespace(),
		},
	}
	return obj
}

func (r *RBAC) buildClusterRole() *rbacv1.ClusterRole {
	obj := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: CSIClusterRoleName,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods"},
				Verbs:     []string{"get", "list", "watch", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumes"},
				Verbs:     []string{"get", "list", "watch", "create", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"persistentvolumeclaims"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"csidrivers"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"storage.k8s.io"},
				Resources: []string{"storageclasses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"listeners.zncdata.dev"},
				Resources: []string{"listenerclasses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"listeners.zncdata.dev"},
				Resources: []string{"listeners"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
		},
	}
	return obj
}

func (r *RBAC) buildClusterRoleBinding() *rbacv1.ClusterRoleBinding {

	obj := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: CSIClusterRoleBindingName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      CSIServiceAccountName,
				Namespace: r.cr.GetNamespace(),
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     CSIClusterRoleName,
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	return obj
}
