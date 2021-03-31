package main

import (
	"context"
	"encoding/json"
	"fmt"

	// goflag "flag"
	flag "github.com/spf13/pflag"

	"github.com/safanaj/k8s-node-role-labeler/pkg/awsutils"
)

const tagPrefix = "node-role.kubernetes.io/"

func main() {
	tagP := flag.String("tag", "", "AAAA")
	iidP := flag.String("instance-id", "", "AAAA")

	flag.Parse()

	ctx := context.TODO()
	cache := awsutils.NewCache(ctx, *tagP)

	if *iidP != "" {
		if *tagP == "" {
			i, err := awsutils.GetEC2InstanceById(ctx, *iidP)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
			jsonBytes, _ := json.MarshalIndent(i, "", "  ")
			fmt.Printf("%s\n", jsonBytes)
		} else {
			role, err := cache.Get(*iidP)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}
			fmt.Printf("%s is a %s\n", *iidP, tagPrefix+role)
		}
	}
}
