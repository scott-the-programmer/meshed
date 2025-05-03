package apps

import (
	"fmt"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/yaml"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func NewBlog(ctx *pulumi.Context, provider *kubernetes.Provider, ns *corev1.Namespace, name string) error {

	appLabels := pulumi.StringMap{
		"app": pulumi.String(name),
	}

	_, err := appsv1.NewDeployment(ctx, fmt.Sprintf("%s-deployment", name), &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: ns.Metadata.Name(),
		},
		Spec: appsv1.DeploymentSpecArgs{
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: appLabels,
			},
			Replicas: pulumi.Int(1),
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels:    appLabels,
					Namespace: ns.Metadata.Name(),
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						corev1.ContainerArgs{
							Name:  pulumi.String(name),
							Image: pulumi.String("ghcr.io/scott-the-programmer/smkiwi/smkiwi:latest"),
						}},
				},
			},
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	_, err = corev1.NewService(ctx, fmt.Sprintf("%s-svc", name), &corev1.ServiceArgs{
		Metadata: metav1.ObjectMetaArgs{
			Name:      pulumi.String(name),
			Namespace: ns.Metadata.Name(),
			Labels: pulumi.StringMap{
				"app":     pulumi.String(name),
				"service": pulumi.String(name),
			},
		},
		Spec: &corev1.ServiceSpecArgs{
			Type: pulumi.String("ClusterIP"),
			Ports: &corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(9080),
					TargetPort: pulumi.Int(80), //Original Target Port
					Protocol:   pulumi.String("TCP"),
				},
			},
			Selector: appLabels,
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	//Remove the ingress
	// _, err = yaml.NewConfigFile(ctx, fmt.Sprintf("%s-deployment", name),
	// 	&yaml.ConfigFileArgs{
	// 		File: fmt.Sprintf("apps/%s.yaml", name),
	// 	},
	// 	pulumi.Provider(provider), pulumi.Parent(ns),
	// )
	// if err != nil {
	// 	return err
	// }

	return nil
}
