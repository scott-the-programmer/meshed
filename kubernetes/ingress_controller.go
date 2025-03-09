package kubernetes

import (
	"meshed/kubernetes/resources"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// NewIngressController creates and configures an NGINX ingress controller
func NewIngressController(ctx *pulumi.Context, provider *kubernetes.Provider, replacer *resources.Replacer) error {
	// Create namespace for ingress controller
	ingressNs, err := corev1.NewNamespace(ctx, "ingress-nginx", &corev1.NamespaceArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name: pulumi.String("ingress-nginx"),
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	// Install NGINX ingress controller using Helm
	_, err = helm.NewRelease(ctx, "ingress-nginx", &helm.ReleaseArgs{
		Chart:     pulumi.String("ingress-nginx"),
		Version:   pulumi.String("4.7.1"),
		Namespace: ingressNs.Metadata.Name().Elem(),
		RepositoryOpts: helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://kubernetes.github.io/ingress-nginx"),
		},
		Values: pulumi.Map{
			"controller": pulumi.Map{
				"service": pulumi.Map{
					"type": pulumi.String("LoadBalancer"),
				},
				"resources": pulumi.Map{
					"requests": pulumi.Map{
						"cpu":    pulumi.String("100m"),
						"memory": pulumi.String("90Mi"),
					},
					"limits": pulumi.Map{
						"cpu":    pulumi.String("200m"),
						"memory": pulumi.String("180Mi"),
					},
				},
			},
		},
		Timeout:     pulumi.Int(600),
		WaitForJobs: pulumi.Bool(true),
		Atomic:      pulumi.Bool(true),
	}, pulumi.Provider(provider), pulumi.Parent(ingressNs))

	return err
}
