package linode

import (
	"encoding/base64"

	"github.com/pulumi/pulumi-linode/sdk/v3/go/linode"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// NewLkeCluster creates a linode kubernetes cluster
func NewLkeCluster(ctx *pulumi.Context, name string, version string) (pulumi.StringOutput, error) {
	cluster, err := linode.NewLkeCluster(ctx, name, &linode.LkeClusterArgs{
		K8sVersion: pulumi.String(version),
		Label:      pulumi.String(name),
		Pools: linode.LkeClusterPoolArray{
			&linode.LkeClusterPoolArgs{
				Count: pulumi.Int(2),
				Type:  pulumi.String("g6-standard-1"),
			},
		},
		Region: pulumi.String("ap-south"),
	})
	if err != nil {
		return pulumi.StringOutput{}, err
	}

	config := cluster.Kubeconfig.ApplyT(func(kubeconfig string) string {
		decoded, _ := base64.StdEncoding.DecodeString(kubeconfig)
		return string(decoded)
	}).(pulumi.StringOutput)
	return config, nil
}
