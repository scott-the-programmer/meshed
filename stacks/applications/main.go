package main

import (
	"meshed/kubernetes"
	"meshed/stacks/applications/apps"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		clusterStack, err := pulumi.NewStackReference(ctx, "scott-the-programmer/meshed/cluster", nil)
		if err != nil {
			return err
		}
		config := clusterStack.GetStringOutput(pulumi.String("kubeconfig"))

		provider, err := kubernetes.NewKubernetesProvider(ctx, config)
		if err != nil {
			return err
		}

		appNS, err := corev1.NewNamespace(ctx, "personal",
			&corev1.NamespaceArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name: pulumi.String("personal"),
					Labels: pulumi.StringMap{
						"istio-injection": pulumi.String("enabled"),
					}},
			}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		err = apps.NewBlog(ctx, provider, appNS, "blog")
		if err != nil {
			return err
		}

		err = apps.NewTermNz(ctx, provider, appNS, "term-nz")
		if err != nil {
			return err
		}

		return nil
	})
}
