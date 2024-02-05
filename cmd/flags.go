package main

import (
	goflag "flag"
	flag "github.com/spf13/pflag"

	"k8s.io/klog/v2"

	"github.com/mjimenezp/k8s-node-role-labeler/pkg/options"
	"github.com/mjimenezp/k8s-node-role-labeler/pkg/reconcilers"
)

type Flags struct {
	version     bool
	labelPrefix string

	opts *options.Options

	// tagKey          string
	// useAwsLifecycle bool

	// fromLabel       string
	// fromLabelPrefix string
	// fromLabelSuffix string
}

func parseFlags() *Flags {
	flags := &Flags{opts: &options.Options{}}

	klog.InitFlags(nil)

	flag.BoolVar(&flags.version, "version", false, "Print version and exit")
	// TODO: the aws-tag-key could be derived/autodiscovered using the cluster name where we are running on
	flag.StringVar(&flags.labelPrefix, "node-label-prefix", reconcilers.LabelPrefix, "Prefix of the node label to tag the role")
	flag.StringVar(&flags.opts.TagKey, "aws-tag-key", "", "AWS tag-key to discover instances belonging the cluster")
	flag.BoolVar(&flags.opts.UseAwsLifecycle, "use-aws-instance-lifecycle", false, "Connect to AWS and inspect instance lifecycle")
	flag.StringVar(&flags.opts.FromLabel, "from-label", "", "Copy role from node label")
	flag.StringVar(&flags.opts.FromLabelPrefix, "from-label-prefix", "", "Add prefix copying from label")
	flag.StringVar(&flags.opts.FromLabelSuffix, "from-label-suffix", "", "Add suffix copying from label")

	flag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	flag.Parse()
	return flags
}
