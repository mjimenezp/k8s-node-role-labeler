package awsutils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	// use aws sdk go v2
	"github.com/aws/aws-sdk-go-v2/aws"
	// awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

const METADATA_AZ_URL = "http://169.254.169.254/latest/meta-data/placement/availability-zone"

func detectCurrentRegion(ctx context.Context) (string, error) {
	var (
		err error
		// body []byte
		body *strings.Builder
		resp *http.Response
	)

	req, _ := http.NewRequestWithContext(
		ctx, http.MethodGet, METADATA_AZ_URL, nil)
	if resp, err = http.DefaultClient.Do(req); err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body = &strings.Builder{}
	if _, err := io.Copy(body, resp.Body); err != nil {
		return "", err
	}
	if body.Len() < 2 {
		return "", fmt.Errorf("Invalid AZ detected: %v", body)
	}
	return body.String()[:body.Len()-1], nil
}

func GetAwsConfigOrDie(ctx context.Context) aws.Config {
	var region string
	if region_, ok := os.LookupEnv("AWS_REGION"); !ok {
		if region_, err := detectCurrentRegion(ctx); err == nil {
			region = region_
		} else {
			panic(err)
		}
	} else {
		region = region_
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithDefaultRegion(region))
	if err != nil {
		panic(err)
	}
	return awsCfg
}

var _ec2svc *ec2.Client

func GetEC2ServiceOrDie(ctx context.Context) *ec2.Client {
	if _ec2svc == nil {
		_ec2svc = ec2.NewFromConfig(GetAwsConfigOrDie(ctx))
	}
	return _ec2svc
}

var (
	WrongReservationsNumberErr error = errors.New("WrongReservationsNumberErr")
	WrongInstancesNumberErr    error = errors.New("WrongInstancesNumberErr")
)

func GetEC2InstanceById(ctx context.Context, id string) (types.Instance, error) {
	ec2svc := GetEC2ServiceOrDie(ctx)
	in := &ec2.DescribeInstancesInput{InstanceIds: []string{id}}
	out, err := ec2svc.DescribeInstances(ctx, in)
	if err != nil {
		return types.Instance{}, err
	}
	if len(out.Reservations) != 1 {
		return types.Instance{}, WrongReservationsNumberErr
	}
	if len(out.Reservations[0].Instances) != 1 {
		return types.Instance{}, WrongInstancesNumberErr
	}
	return out.Reservations[0].Instances[0], nil
}

// This is the logic about howto decide the role of a node
// TODO: move those -worker strings outside
func ComputeNodeLabelKeySuffixForInstance(instance types.Instance) string {
	t := instance.Placement.Tenancy
	lc := instance.InstanceLifecycle
	// TODO: eventually consider also scheduled lifecycle
	if t == "default" {
		if lc == "" {
			return "ondemand-worker"
		} else {
			return "spot-worker"
		}
	} else {
		// TODO: evaluate also host and dedicated meaning
		return "dedicated-worker"
	}
}
