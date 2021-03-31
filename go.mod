module github.com/safanaj/k8s-node-role-labeler

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.3.0
	github.com/aws/aws-sdk-go-v2/config v1.1.3
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.2.0
	github.com/go-logr/logr v0.3.0
	github.com/spf13/pflag v1.0.5
	k8s.io/api v0.20.2
	k8s.io/klog/v2 v2.4.0
	sigs.k8s.io/controller-runtime v0.8.3
)
