package workflow

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	v1beta2 "k8s.io/api/apps/v1beta2"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apixv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apixv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/ian-howell/airshipctl/pkg/environment"
	wfenv "github.com/ian-howell/airshipctl/pkg/workflow/environment"
)

const (
	argoNamespace = "argo"
)

var (
	manifestPath string
)

type workflowInitCmd struct {
	out        io.Writer
	kubeclient kubernetes.Interface
	crdclient  apixv1beta1client.ApiextensionsV1beta1Interface
}

// NewWorkflowInitCommand is a command for bootstrapping a kubernetes cluster with the necessary components for Argo workflows
func NewWorkflowInitCommand(rootSettings *environment.AirshipCTLSettings) *cobra.Command {

	workflowInitCommand := &cobra.Command{
		Use:   "init [flags]",
		Short: "bootstraps the kubernetes cluster with the Workflow CRDs and controller",
		Run: func(cmd *cobra.Command, args []string) {
			out := cmd.OutOrStdout()
			wfSettings, ok := rootSettings.PluginSettings[PluginSettingsID].(*wfenv.Settings)
			if !ok {
				fmt.Fprintf(out, "settings for %s were not registered\n", PluginSettingsID)
				return
			}

			workflowInit := &workflowInitCmd{
				out:        out,
				kubeclient: wfSettings.KubeClient,
				crdclient:  wfSettings.CRDClient.ApiextensionsV1beta1(),
			}

			fmt.Fprintf(out, "Creating namespace \"%s\"\n", argoNamespace)
			_, err := workflowInit.kubeclient.CoreV1().Namespaces().Create(&v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{Name: "argo"},
			})
			if err != nil {
				if errors.IsAlreadyExists(err) {
					fmt.Fprintf(out, "Namespace \"%s\" already exists\n", argoNamespace)
				} else {
					fmt.Fprintf(out, "Could not create namespace \"%s\": %s\n", argoNamespace, err.Error())
					return
				}
			}

			if manifestPath == "" {
				if err := workflowInit.createDefaultObjects(); err != nil {
					fmt.Fprintf(out, "Could not create default objects: %s\n", err.Error())
					return
				}
			} else {
				workflowInit.createCustomObjects(manifestPath)
			}

		},
	}

	workflowInitCommand.PersistentFlags().StringVar(&manifestPath, "manifest", "", "path to a YAML manifest containing definitions of objects needed for Argo workflows")
	return workflowInitCommand
}

func (wfInit *workflowInitCmd) createDefaultObjects() error {
	//TODO(howell): Clean up the repetitive code
	fmt.Fprintf(wfInit.out, "Registering Workflow CRD\n")
	if err := wfInit.registerDefaultWorkflow(); err != nil {
		if errors.IsAlreadyExists(err) {
			fmt.Fprintf(wfInit.out, "Workflow CRD already registered\n")
		} else {
			return fmt.Errorf("Could not register Workflow CRD: %s", err.Error())
		}
	}

	fmt.Fprintf(wfInit.out, "Creating argo ServiceAccount\n")
	if err := wfInit.createArgoServiceAccount(); err != nil {
		if errors.IsAlreadyExists(err) {
			fmt.Fprintf(wfInit.out, "argo ServiceAccount already exists\n")
		} else {
			return fmt.Errorf("Could not create argo ServiceAccount: %s", err.Error())
		}
	}

	fmt.Fprintf(wfInit.out, "Creating argo admin ClusterRole\n")
	if err := wfInit.createArgoAdminClusterRole(); err != nil {
		if errors.IsAlreadyExists(err) {
			fmt.Fprintf(wfInit.out, "argo admin ClusterRole already exists\n")
		} else {
			return fmt.Errorf("Could not create argo admin ClusterRole: %s", err.Error())
		}
	}

	fmt.Fprintf(wfInit.out, "Creating argo edit ClusterRole\n")
	if err := wfInit.createArgoEditClusterRole(); err != nil {
		if errors.IsAlreadyExists(err) {
			fmt.Fprintf(wfInit.out, "argo edit ClusterRole already exists\n")
		} else {
			return fmt.Errorf("Could not create argo edit ClusterRole: %s", err.Error())
		}
	}

	fmt.Fprintf(wfInit.out, "Creating argo view ClusterRole\n")
	if err := wfInit.createArgoViewClusterRole(); err != nil {
		if errors.IsAlreadyExists(err) {
			fmt.Fprintf(wfInit.out, "argo view ClusterRole already exists\n")
		} else {
			return fmt.Errorf("Could not create argo view ClusterRole: %s", err.Error())
		}
	}

	fmt.Fprintf(wfInit.out, "Creating argo ClusterRole\n")
	if err := wfInit.createArgoClusterRole(); err != nil {
		if errors.IsAlreadyExists(err) {
			fmt.Fprintf(wfInit.out, "argo ClusterRole already exists\n")
		} else {
			return fmt.Errorf("Could not create argo ClusterRole: %s", err.Error())
		}
	}

	fmt.Fprintf(wfInit.out, "Creating argo ClusterRoleBinding\n")
	if err := wfInit.createArgoClusterRoleBinding(); err != nil {
		if errors.IsAlreadyExists(err) {
			fmt.Fprintf(wfInit.out, "argo ClusterRoleBinding already exists\n")
		} else {
			return fmt.Errorf("Could not create argo ClusterRoleBinding: %s", err.Error())
		}
	}

	fmt.Fprintf(wfInit.out, "Creating argo ConfigMap\n")
	if err := wfInit.createArgoConfigMap(); err != nil {
		if errors.IsAlreadyExists(err) {
			fmt.Fprintf(wfInit.out, "argo ConfigMap already exists\n")
		} else {
			return fmt.Errorf("Could not create argo ConfigMap: %s", err.Error())
		}
	}

	fmt.Fprintf(wfInit.out, "Creating argo Deployment\n")
	if err := wfInit.createArgoDeployment(); err != nil {
		if errors.IsAlreadyExists(err) {
			fmt.Fprintf(wfInit.out, "argo Deployment already exists\n")
		} else {
			return fmt.Errorf("Could not create argo Deployment: %s", err.Error())
		}
	}

	return nil
}

func (wfInit *workflowInitCmd) createCustomObjects(manifestPath string) {
	//TODO
}

func (wfInit *workflowInitCmd) registerDefaultWorkflow() error {
	apixClient := wfInit.crdclient
	crds := apixClient.CustomResourceDefinitions()
	workflowCRD := &apixv1beta1.CustomResourceDefinition{
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
	}
	_, err := crds.Create(workflowCRD)
	return err
}

func (wfInit *workflowInitCmd) createArgoServiceAccount() error {
	_, err := wfInit.kubeclient.CoreV1().ServiceAccounts(argoNamespace).Create(&v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "argo",
			Namespace: argoNamespace,
		},
	})
	return err
}

func (wfInit *workflowInitCmd) createArgoAdminClusterRole() error {
	_, err := wfInit.kubeclient.RbacV1().ClusterRoles().Create(&rbacv1.ClusterRole{
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

func (wfInit *workflowInitCmd) createArgoEditClusterRole() error {
	_, err := wfInit.kubeclient.RbacV1().ClusterRoles().Create(&rbacv1.ClusterRole{
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

func (wfInit *workflowInitCmd) createArgoViewClusterRole() error {
	_, err := wfInit.kubeclient.RbacV1().ClusterRoles().Create(&rbacv1.ClusterRole{
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

func (wfInit *workflowInitCmd) createArgoClusterRole() error {
	_, err := wfInit.kubeclient.RbacV1().ClusterRoles().Create(&rbacv1.ClusterRole{
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

func (wfInit *workflowInitCmd) createArgoClusterRoleBinding() error {
	_, err := wfInit.kubeclient.RbacV1().ClusterRoleBindings().Create(&rbacv1.ClusterRoleBinding{
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

func (wfInit *workflowInitCmd) createArgoConfigMap() error {
	_, err := wfInit.kubeclient.CoreV1().ConfigMaps(argoNamespace).Create(&v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "workflow-controller-configmap",
			Namespace: argoNamespace,
		},
	})
	return err
}

func (wfInit *workflowInitCmd) createArgoDeployment() error {
	_, err := wfInit.kubeclient.AppsV1beta2().Deployments(argoNamespace).Create(&v1beta2.Deployment{
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
