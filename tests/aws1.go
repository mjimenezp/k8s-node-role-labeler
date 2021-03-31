package main

import (
	"context"
	"encoding/json"
	"fmt"

	// goflag "flag"
	flag "github.com/spf13/pflag"

	"github.com/safanaj/k8s-node-role-labeler/pkg/awsutils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const tagPrefix = "node-role.kubernetes.io/"

func main() {
	tagP := flag.String("tag", "", "AAAA")
	iidP := flag.String("instance-id", "", "AAAA")

	flag.Parse()

	ctx := context.TODO()
	ec2svc := awsutils.GetEC2ServiceOrDie(ctx)
	cache := make(map[string]string)

	if *tagP != "" {
		p := ec2.NewDescribeInstancesPaginator(ec2svc, &ec2.DescribeInstancesInput{
			Filters: []ec2types.Filter{{
				Name: aws.String("tag-key"), Values: []string{*tagP},
			}},
		})

		for {
			if !p.HasMorePages() {
				break
			}
			out, err := p.NextPage(ctx)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}

			for _, r := range out.Reservations {
				for _, i := range r.Instances {
					id := *i.InstanceId
					t := i.Placement.Tenancy
					lc := i.InstanceLifecycle
					tagSuffix := ""
					if t == "default" {
						if lc == "" {
							tagSuffix = "ondemand-worker"
						} else {
							tagSuffix = "spot-worker"
						}
					} else {
						tagSuffix = "dedicated-worker"
					}
					cache[id] = tagPrefix + tagSuffix
				}
			}
		}
		if *iidP == "" {
			jsonBytes, _ := json.MarshalIndent(cache, "", "  ")
			fmt.Printf("%s\n", jsonBytes)
		}
	}

	if *iidP != "" {
		if *tagP == "" {
			out, err := ec2svc.DescribeInstances(ctx, &ec2.DescribeInstancesInput{InstanceIds: []string{*iidP}})
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
			if len(out.Reservations) != 1 {
				fmt.Printf("DescribeInstances returned wrong number in reservations\n")
				return
			}
			if len(out.Reservations[0].Instances) != 1 {
				fmt.Printf("DescribeInstances returned wrong number of instances\n")
				return
			}
			jsonBytes, _ := json.MarshalIndent(out.Reservations[0].Instances[0], "", "  ")
			fmt.Printf("%s\n", jsonBytes)
		} else {
			fmt.Printf("%s is a %s\n", *iidP, cache[*iidP])
		}
	}
}
