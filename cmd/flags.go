package main

import (
	goflag "flag"
	flag "github.com/spf13/pflag"

	"k8s.io/klog/v2"

	"github.com/safanaj/k8s-node-role-labeler/pkg/reconcilers"
)

type Flags struct {
	version     bool
	tagKey      string
	labelPrefix string
}

func parseFlags() *Flags {
	flags := &Flags{}

	klog.InitFlags(nil)

	flag.BoolVar(&flags.version, "version", false, "Print version and exit")
	// TODO: the aws-tag-key could be derived/autodiscovered using the cluster name where we are running on
	flag.StringVar(&flags.tagKey, "aws-tag-key", "", "AWS tag-key to discover instances belonging the cluster")
	flag.StringVar(&flags.labelPrefix, "node-label-prefix", reconcilers.LabelPrefix, "Prefix of the node label to tag the role")

	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
	return flags
}
