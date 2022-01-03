package scaleway

import (
	"github.com/lbrlabs/pulumi-scaleway/sdk/go/scaleway"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// NewLkeCluster creates a linode kubernetes cluster
func NewKapsuleCluster(ctx *pulumi.Context, name string, version string) (pulumi.StringOutput, error) {
	cluster, err := scaleway.NewKubernetesCluster(ctx, name, &scaleway.KubernetesClusterArgs{
		DeleteAdditionalResources: pulumi.Bool(true),
		AutoUpgrade: scaleway.KubernetesClusterAutoUpgradeArgs{
			Enable:                     pulumi.Bool(false),
			MaintenanceWindowDay:       pulumi.String("monday"),
			MaintenanceWindowStartHour: pulumi.Int(0),
		},
		Name:    pulumi.String(name),
		Type:    pulumi.String("kapsule"),
		Version: pulumi.String(version),
		Cni:     pulumi.String("calico"),
	})
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	_, err = scaleway.NewKubernetesNodePool(ctx, name, &scaleway.KubernetesNodePoolArgs{
		ClusterId:        cluster.ID(),
		ContainerRuntime: nil,
		KubeletArgs:      nil,
		MaxSize:          pulumi.IntPtr(3),
		MinSize:          pulumi.IntPtr(1),
		Size:             pulumi.Int(3),
		Name:             pulumi.String(name),
		NodeType:         pulumi.String("DEV1-M"),
	})
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	return pulumi.StringOutput(cluster.Kubeconfigs.Index(pulumi.Int(0)).ConfigFile()), nil
}
