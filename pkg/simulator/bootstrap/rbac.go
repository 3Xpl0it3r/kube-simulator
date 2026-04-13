package bootstrap

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"3Xpl0it3r.com/kube-simulator/pkg/apiclient"

	rbac "k8s.io/api/rbac/v1"
	clientset "k8s.io/client-go/kubernetes"
)

// configure some rbac for allow use to execute "kubeadm join' to join new node
// blew code is from official repo of kubeadm

const (
	NodeKubeletBootstrap            = "kubeadm:kubelet-bootstrap"
	NodeBootstrapperClusterRoleName = "system:node-bootstrapper"
	GetNodesClusterRoleName         = "kubeadm:get-nodes"

	CSRAutoApprovalClusterRoleName = "system:certificates.k8s.io:certificatesigningrequests:nodeclient"
	NodesGroup                     = "system:nodes"

	NodeBootstrapTokenAuthGroup                          = "system:bootstrappers:kubeadm:default-node-token"
	NodeSelfCSRAutoApprovalClusterRoleName               = "system:certificates.k8s.io:certificatesigningrequests:selfnodeclient"
	NodeAutoApproveBootstrapClusterRoleBinding           = "kubeadm:node-autoapprove-bootstrap"
	NodeAutoApproveCertificateRotationClusterRoleBinding = "kubeadm:node-autoapprove-certificate-rotation"
)

// AllowBootstrapTokensToPostCSRs creates RBAC rules in a way the makes Node Bootstrap Tokens able to post CSRs
func AllowBootstrapTokensToPostCSRs(client clientset.Interface) error {

	return apiclient.CreateOrUpdateClusterRoleBinding(client, &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: NodeKubeletBootstrap,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
			Name:     NodeBootstrapperClusterRoleName,
		},
		Subjects: []rbac.Subject{
			{
				Kind: rbac.GroupKind,
				Name: NodeBootstrapTokenAuthGroup,
			},
		},
	})
}

// AllowBoostrapTokensToGetNodes creates RBAC rules to allow Node Bootstrap Tokens to list nodes
func AllowBoostrapTokensToGetNodes(client clientset.Interface) error {

	if err := apiclient.CreateOrUpdateClusterRole(client, &rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetNodesClusterRoleName,
			Namespace: metav1.NamespaceSystem,
		},
		Rules: []rbac.PolicyRule{
			{
				Verbs:     []string{"get"},
				APIGroups: []string{""},
				Resources: []string{"nodes"},
			},
		},
	}); err != nil {
		return err
	}

	return apiclient.CreateOrUpdateClusterRoleBinding(client, &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      GetNodesClusterRoleName,
			Namespace: metav1.NamespaceSystem,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
			Name:     GetNodesClusterRoleName,
		},
		Subjects: []rbac.Subject{
			{
				Kind: rbac.GroupKind,
				Name: NodeBootstrapTokenAuthGroup,
			},
		},
	})
}

// AutoApproveNodeBootstrapTokens creates RBAC rules in a way that makes Node Bootstrap Tokens' CSR auto-approved by the csrapprover controller
func AutoApproveNodeBootstrapTokens(client clientset.Interface) error {

	// Always create this kubeadm-specific binding though
	return apiclient.CreateOrUpdateClusterRoleBinding(client, &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: NodeAutoApproveBootstrapClusterRoleBinding,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
			Name:     CSRAutoApprovalClusterRoleName,
		},
		Subjects: []rbac.Subject{
			{
				Kind: "Group",
				Name: NodeBootstrapTokenAuthGroup,
			},
		},
	})
}

// AutoApproveNodeCertificateRotation creates RBAC rules in a way that makes Node certificate rotation CSR auto-approved by the csrapprover controller
func AutoApproveNodeCertificateRotation(client clientset.Interface) error {

	return apiclient.CreateOrUpdateClusterRoleBinding(client, &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: NodeAutoApproveCertificateRotationClusterRoleBinding,
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
			Name:     NodeSelfCSRAutoApprovalClusterRoleName,
		},
		Subjects: []rbac.Subject{
			{
				Kind: "Group",
				Name: NodesGroup,
			},
		},
	})
}
