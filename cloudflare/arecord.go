package cloudflare

import (
	"fmt"

	"github.com/pulumi/pulumi-cloudflare/sdk/v3/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ARecordArgs struct type
type ARecordArgs struct {
	Source  string
	Target  pulumi.StringOutput
	ZoneID  string
	Proxied bool
}

// NewARecord instantiates an ARecord
func NewARecord(ctx *pulumi.Context, args ARecordArgs) (*cloudflare.Record, error) {
	record, err := cloudflare.NewRecord(ctx, fmt.Sprintf("%s-%s", "a", args.Source), &cloudflare.RecordArgs{
		Name:    pulumi.String(args.Source),
		ZoneId:  pulumi.String(args.ZoneID),
		Type:    pulumi.String("A"),
		Value:   args.Target,
		Proxied: pulumi.Bool(args.Proxied),
	})
	if err != nil {
		return nil, err
	}

	return record, nil
}
