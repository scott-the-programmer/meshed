package kubernetes

import (
	"meshed/kubernetes/resources"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func NewCertManager(ctx *pulumi.Context, gw *helm.Release, provider *kubernetes.Provider, replacer *resources.Replacer) (*helm.Release, error) {

	ns, err := corev1.NewNamespace(ctx, "cert-manager", &corev1.NamespaceArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name: pulumi.String("cert-manager"),
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}

	certMan, err := helm.NewRelease(ctx, "cert-manager", &helm.ReleaseArgs{
		Chart: pulumi.String("cert-manager"),
		Name:  pulumi.String("cert-manager"),
		Values: pulumi.Map{
			"installCRDs": pulumi.Bool(true),
		},
		Namespace: ns.Metadata.Name().Elem().ToStringOutput(),
		RepositoryOpts: helm.RepositoryOptsArgs{
			Repo: pulumi.String("https://charts.jetstack.io"),
		},
		WaitForJobs: pulumi.Bool(true),
		Timeout:     pulumi.Int(600),
	}, pulumi.Provider(provider), pulumi.Parent(gw), pulumi.DeleteBeforeReplace(true))
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return certMan, nil
}
