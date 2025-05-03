package apps

import (
	"fmt"

	"github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes"
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v3/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type CloudflaredArgs struct {
	TunnelSecretName pulumi.StringInput
	TunnelSecretKey  pulumi.StringInput
	Image            pulumi.StringPtrInput
	Subdomain        pulumi.StringInput
	Domain          pulumi.StringInput
}

func NewCloudflared(ctx *pulumi.Context, provider *kubernetes.Provider, ns *corev1.Namespace, name string, args *CloudflaredArgs) error {
	appLabels := pulumi.StringMap{
		"app": pulumi.String(name),
	}

	image := pulumi.StringPtr("cloudflare/cloudflared:latest")
	if args.Image != nil {
		image = pulumi.Sprintf("%v", args.Image).Ptr()
	}

	deployment, err := appsv1.NewDeployment(ctx, fmt.Sprintf("%s-deployment", name), &appsv1.DeploymentArgs{
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
							Image: image,
							Args: pulumi.StringArray{
								pulumi.String("tunnel"),
								pulumi.String("run"),
								pulumi.String("--token"),
								pulumi.Sprintf("%s.%s", args.Subdomain, args.Domain),
							},
							Env: corev1.EnvVarArray{
								corev1.EnvVarArgs{
									Name: pulumi.String("TUNNEL_TOKEN"),
									ValueFrom: &corev1.EnvVarSourceArgs{
										SecretKeyRef: &corev1.SecretKeySelectorArgs{
											LocalObjectReference: &corev1.LocalObjectReferenceArgs{
												Name: args.TunnelSecretName,
											},
											Key: args.TunnelSecretKey,
										},
									},
								},
							},
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
					Port:       pulumi.Int(8080), // Cloudflared listens on 8080 by default
					TargetPort: pulumi.Int(8080),
					Protocol:   pulumi.String("TCP"),
				},
			},
			Selector: appLabels,
		},
	}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{deployment}))
	if err != nil {
		return err
	}

	return nil
}
