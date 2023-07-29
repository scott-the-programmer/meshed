package main

import (
	"meshed/kubernetes"

	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
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

		ns, err := corev1.NewNamespace(ctx, "monitoring",
			&corev1.NamespaceArgs{
				Metadata: metav1.ObjectMetaArgs{
					Name: pulumi.String("monitoring"),
				}}, pulumi.Provider(provider))
		if err != nil {
			return err
		}

		_, err = yaml.NewConfigFile(ctx, "monitoring-deployment",
			&yaml.ConfigFileArgs{
				File: "grafana.yaml",
			},
			pulumi.Provider(provider), pulumi.Parent(ns))
		if err != nil {
			return err
		}

		return nil
	})
}
