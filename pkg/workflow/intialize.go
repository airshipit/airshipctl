package workflow

import (
	"fmt"
	"io"

	v1beta2 "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apixv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apixv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/ian-howell/airshipctl/pkg/workflow/environment"
)

const (
	argoNamespace = "argo"
)

// Initialize creates the argo namespace and all of the required kubernetes
// objects for argo to function
func Initialize(out io.Writer, settings *environment.Settings, manifestPath string) error {
	kubeClient := settings.KubeClient
	crdClient := settings.CRDClient

	if err := createNamespace(out, kubeClient); err != nil {
		return err
	}

	if manifestPath == "" {
		if err := createDefaultObjects(out, kubeClient, crdClient); err != nil {
			return fmt.Errorf("could not create default objects: %s", err.Error())
		}
	} else {
		if err := createCustomObjects(manifestPath); err != nil {
			return fmt.Errorf("could not create objects: %s", err.Error())
		}
	}

	return nil
}

func createNamespace(out io.Writer, kubeClient kubernetes.Interface) error {
	fmt.Fprintf(out, "Creating namespace %s\n", argoNamespace)
	_, err := kubeClient.CoreV1().Namespaces().Create(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "argo"}})
	return handleCreateError(out, fmt.Sprintf("namespace %s", argoNamespace), err)
}

func createDefaultObjects(out io.Writer, kubeClient kubernetes.Interface, crdClient apixv1beta1client.Interface) error {
	fmt.Fprintf(out, "Registering Workflow CRD\n")
	if err := handleCreateError(out, "Workflow CRD", registerDefaultWorkflow(crdClient)); err != nil {
		return err
	}

	fmt.Fprintf(out, "Creating argo ServiceAccount\n")
	if err := handleCreateError(out, "argo ServiceAccount", createArgoServiceAccount(kubeClient)); err != nil {
		return err
	}

	fmt.Fprintf(out, "Creating argo admin ClusterRole\n")
	if err := handleCreateError(out, "argo admin ClusterRole", createArgoAdminClusterRole(kubeClient)); err != nil {
		return err
	}

	fmt.Fprintf(out, "Creating argo edit ClusterRole\n")
	if err := handleCreateError(out, "argo edit ClusterRole", createArgoEditClusterRole(kubeClient)); err != nil {
		return err
	}

	fmt.Fprintf(out, "Creating argo view ClusterRole\n")
	if err := handleCreateError(out, "argo view ClusterRole", createArgoViewClusterRole(kubeClient)); err != nil {
		return err
	}

	fmt.Fprintf(out, "Creating argo ClusterRole\n")
	if err := handleCreateError(out, "argo ClusterRole", createArgoClusterRole(kubeClient)); err != nil {
		return err
	}

	fmt.Fprintf(out, "Creating argo ClusterRoleBinding\n")
	if err := handleCreateError(out, "argo ClusterRoleBinding", createArgoClusterRoleBinding(kubeClient)); err != nil {
		return err
	}

	fmt.Fprintf(out, "Creating argo ConfigMap\n")
	if err := handleCreateError(out, "argo ConfigMap", createArgoConfigMap(kubeClient)); err != nil {
		return err
	}

	fmt.Fprintf(out, "Creating argo Deployment\n")
	if err := handleCreateError(out, "argo Deployment", createArgoDeployment(kubeClient)); err != nil {
		return err
	}

	return nil
}

func createCustomObjects(manifestPath string) error {
	//TODO
	return nil
}

func registerDefaultWorkflow(crdClient apixv1beta1client.Interface) error {
	_, err := crdClient.ApiextensionsV1beta1().CustomResourceDefinitions().
		Create(&apixv1beta1.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: "workflows.argoproj.io",
			},
			Spec: apixv1beta1.CustomResourceDefinitionSpec{
				Group:   "argoproj.io",
				Version: "v1alpha1",
				Versions: []apixv1beta1.CustomResourceDefinitionVersion{{
					Name:    "v1alpha1",
					Served:  true,
					Storage: true,
				}},
				Names: apixv1beta1.CustomResourceDefinitionNames{
					Plural: "workflows",
					Kind:   "Workflow",
				},
				Scope: apixv1beta1.NamespaceScoped,
			},
		})
	return err
}

func createArgoServiceAccount(kubeClient kubernetes.Interface) error {
	_, err := kubeClient.CoreV1().ServiceAccounts(argoNamespace).
		Create(&v1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "argo",
				Namespace: argoNamespace,
			},
		})
	return err
}

func createArgoAdminClusterRole(kubeClient kubernetes.Interface) error {
	_, err := kubeClient.RbacV1().ClusterRoles().
		Create(&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: "argo-aggregate-to-admin",
				Labels: map[string]string{
					"rbac.authorization.k8s.io/aggregate-to-admin": "true",
				},
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{
						"argoproj.io",
					},
					Resources: []string{
						"workflows",
						"workflows/finalizers",
					},
					Verbs: []string{
						"create",
						"delete",
						"deletecollection",
						"get",
						"list",
						"patch",
						"update",
						"watch",
					},
				},
			},
		})
	return err
}

func createArgoEditClusterRole(kubeClient kubernetes.Interface) error {
	_, err := kubeClient.RbacV1().ClusterRoles().
		Create(&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: "argo-aggregate-to-edit",
				Labels: map[string]string{
					"rbac.authorization.k8s.io/aggregate-to-edit": "true",
				},
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{
						"argoproj.io",
					},
					Resources: []string{
						"workflows",
						"workflows/finalizers",
					},
					Verbs: []string{
						"create",
						"delete",
						"deletecollection",
						"get",
						"list",
						"patch",
						"update",
						"watch",
					},
				},
			},
		})
	return err
}

func createArgoViewClusterRole(kubeClient kubernetes.Interface) error {
	_, err := kubeClient.RbacV1().ClusterRoles().
		Create(&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{
				Name: "argo-aggregate-to-view",
				Labels: map[string]string{
					"rbac.authorization.k8s.io/aggregate-to-view": "true",
				},
			},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{
						"argoproj.io",
					},
					Resources: []string{
						"workflows",
						"workflows/finalizers",
					},
					Verbs: []string{
						"get",
						"list",
						"watch",
					},
				},
			},
		})
	return err
}

func createArgoClusterRole(kubeClient kubernetes.Interface) error {
	_, err := kubeClient.RbacV1().ClusterRoles().
		Create(&rbacv1.ClusterRole{
			ObjectMeta: metav1.ObjectMeta{Name: "argo-cluster-role"},
			Rules: []rbacv1.PolicyRule{
				{
					APIGroups: []string{
						"",
					},
					Resources: []string{
						"pods",
						"pods/exec",
					},
					Verbs: []string{
						"create",
						"delete",
						"get",
						"list",
						"patch",
						"update",
						"watch",
					},
				},
				{
					APIGroups: []string{
						"",
					},
					Resources: []string{
						"configmaps",
					},
					Verbs: []string{
						"get",
						"list",
						"watch",
					},
				},
				{
					APIGroups: []string{
						"",
					},
					Resources: []string{
						"persistentvolumeclaims",
					},
					Verbs: []string{
						"create",
						"delete",
					},
				},
				{
					APIGroups: []string{
						"argoproj.io",
					},
					Resources: []string{
						"workflows",
						"workflows/finalizers",
					},
					Verbs: []string{
						"delete",
						"get",
						"list",
						"patch",
						"update",
						"watch",
					},
				},
			},
		})
	return err
}

func createArgoClusterRoleBinding(kubeClient kubernetes.Interface) error {
	_, err := kubeClient.RbacV1().ClusterRoleBindings().
		Create(&rbacv1.ClusterRoleBinding{
			ObjectMeta: metav1.ObjectMeta{Name: "argo-binding"},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     "argo-cluster-role",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind:      "ServiceAccount",
					Name:      "argo",
					Namespace: argoNamespace,
				},
			},
		})

	return err
}

func createArgoConfigMap(kubeClient kubernetes.Interface) error {
	_, err := kubeClient.CoreV1().ConfigMaps(argoNamespace).
		Create(&v1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workflow-controller-configmap",
				Namespace: argoNamespace,
			},
		})
	return err
}

func createArgoDeployment(kubeClient kubernetes.Interface) error {
	_, err := kubeClient.AppsV1beta2().Deployments(argoNamespace).
		Create(&v1beta2.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "workflow-controller",
				Namespace: argoNamespace,
			},
			Spec: v1beta2.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": "workflow-controller",
					},
				},
				Template: v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": "workflow-controller",
						},
					},
					Spec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Command: []string{
									"workflow-controller",
								},
								Args: []string{
									"--configmap",
									"workflow-controller-configmap",
									"--executor-image",
									"argoproj/argoexec:v2.2.1", // TODO(howell): Remove this hardcoded value
								},
								Image: "argoproj/argoexec:v2.2.1", // TODO(howell): Remove this hardcoded value
								Name:  "workflow-controller",
							},
						},
						ServiceAccountName: "argo",
					},
				},
			},
		})
	return err
}

func handleCreateError(out io.Writer, resourceName string, err error) error {
	if err == nil {
		return nil
	}
	if errors.IsAlreadyExists(err) {
		fmt.Fprintf(out, "%s already exists\n", resourceName)
		return nil
	}
	return fmt.Errorf("Could not create %s: %s", resourceName, err.Error())
}
