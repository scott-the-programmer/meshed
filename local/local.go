package local

import (
	"os"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Returns the current kubeconfig
func NewLocalCluster(ctx *pulumi.Context) (pulumi.StringOutput, error) {
	home := os.Getenv("HOME")
	kubeConfig, err := os.ReadFile(home + "/.kube/config")
	if err != nil {
		return pulumi.StringOutput{}, err
	}
	pulumiStr := pulumi.String(string(kubeConfig))
	return pulumiStr.ToStringOutput(), err
}
