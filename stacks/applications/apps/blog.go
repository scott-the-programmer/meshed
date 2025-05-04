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

type BlogArgs struct {
	Cloudflared *CloudflaredArgs
}

func NewBlog(ctx *pulumi.Context, provider *kubernetes.Provider, ns *corev1.Namespace, name string, args *BlogArgs) error {

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
			Type: pulumi.String("ClusterIP"), // Change to ClusterIP
			Ports: &corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Port:       pulumi.Int(80),       // Change to port 80
					TargetPort: pulumi.Int(9080), //Original Target Port
					Protocol:   pulumi.String("TCP"),
				},
			},
			Selector: appLabels,
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return err
	}

	if args.Cloudflared != nil {
		cfName := fmt.Sprintf("%s-cf", name)
		image := pulumi.StringPtr("cloudflare/cloudflared:latest")
		if args.Cloudflared.Image != nil {
			image = pulumi.Sprintf("%v", args.Cloudflared.Image).ToStringPtrOutput()
		}

		cfDeployment, err := appsv1.NewDeployment(ctx, fmt.Sprintf("%s-cf-deployment", name), &appsv1.DeploymentArgs{
			Metadata: &metav1.ObjectMetaArgs{
				Namespace: ns.Metadata.Name(),
			},
			Spec: appsv1.DeploymentSpecArgs{
				Selector: &metav1.LabelSelectorArgs{
					MatchLabels: pulumi.StringMap{
						"app": pulumi.String(cfName),
					},
				},
				Replicas: pulumi.Int(1),
				Template: &corev1.PodTemplateSpecArgs{
					Metadata: &metav1.ObjectMetaArgs{
						Labels: pulumi.StringMap{
							"app":     pulumi.String(cfName),
							"service": pulumi.String(cfName),
						},
						Namespace: ns.Metadata.Name(),
					},
					Spec: &corev1.PodSpecArgs{
						Containers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name:  pulumi.String(cfName),
								Image: image,
								Args: pulumi.StringArray{
									pulumi.String("tunnel"),
									pulumi.String("run"),
									pulumi.String("--token"),
									pulumi.Sprintf("%s.%s", args.Cloudflared.Subdomain, args.Cloudflared.Domain),
								},
								Env: corev1.EnvVarArray{
									corev1.EnvVarArgs{
										Name: pulumi.String("TUNNEL_TOKEN"),
										ValueFrom: &corev1.EnvVarSourceArgs{
											SecretKeyRef: &corev1.SecretKeySelectorArgs{
												Name: args.Cloudflared.TunnelSecretName,
												Key:  args.Cloudflared.TunnelSecretKey,
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

		_, err = corev1.NewService(ctx, fmt.Sprintf("%s-cf-svc", name), &corev1.ServiceArgs{
			Metadata: metav1.ObjectMetaArgs{
				Name:      pulumi.String(cfName),
				Namespace: ns.Metadata.Name(),
				Labels: pulumi.StringMap{
					"app":     pulumi.String(cfName),
					"service": pulumi.String(cfName),
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
				Selector: pulumi.StringMap{
					"app": pulumi.String(cfName),
				},
			},
		}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{cfDeployment}))
		if err != nil {
			return err
		}
	}

	return nil
}
