package template

import (
	"github.com/kubesphere/kubeeye/pkg/constant"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetClusterRoleTemplate() *rbacv1.ClusterRole {

	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubeeye-inspect-role",
		},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""},
			Resources: []string{"configmaps", "pods", "server", "events", "services", "nodes", "namespaces"},
			Verbs:     []string{"list", "get", "watch"},
		},
			{
				APIGroups: []string{""},
				Resources: []string{"configmaps"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{"apps"},
				Resources: []string{"deployments", "daemonsets", "statefulsets"},
				Verbs:     []string{"list", "get", "watch"},
			},
			{
				APIGroups: []string{"batch"},
				Resources: []string{"jobs", "cronjobs"},
				Verbs:     []string{"list", "get", "watch"},
			},
			{
				APIGroups: []string{"rbac.authorization.k8s.io"},
				Resources: []string{"roles", "clusterroles"},
				Verbs:     []string{"list", "get", "watch"},
			},
		},
	}
}

func GetClusterRoleBindingTemplate() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kubeeye-inspect-rolebinding",
		},
		Subjects: []rbacv1.Subject{
			{Kind: "ServiceAccount", Name: "kubeeye-inspect-job", Namespace: constant.DefaultNamespace},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "kubeeye-inspect-role",
		},
	}
}

func GetServiceAccountTemplate() *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubeeye-inspect-job",
			Namespace: constant.DefaultNamespace,
		},
	}
}
