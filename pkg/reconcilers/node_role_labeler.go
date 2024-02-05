package reconcilers

import (
	"context"
	"errors"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/mjimenezp/k8s-node-role-labeler/pkg/awsutils"
	"github.com/mjimenezp/k8s-node-role-labeler/pkg/options"
)

var LabelPrefix string = "node-role.kubernetes.io/"

var LabelSourceUndefinedErr error = errors.New("LabelSourceUndefinedErr")

type nodeRoleLabelReconciler struct {
	// reconciler is able to retrieve objects from the APIServer.
	client.Client
	logr.Logger

	opts  *options.Options
	cache *awsutils.Cache
}

func (r *nodeRoleLabelReconciler) InjectClient(c client.Client) error {
	r.Client = c
	return nil
}

func (r *nodeRoleLabelReconciler) InjectLogger(l logr.Logger) error {
	r.Logger = l
	return nil
}

func NewNodeRoleLabelReconciler(cache *awsutils.Cache, opts *options.Options) reconcile.Reconciler {
	return &nodeRoleLabelReconciler{cache: cache, opts: opts}
}

func (r *nodeRoleLabelReconciler) Reconcile(ctx context.Context, o reconcile.Request) (reconcile.Result, error) {
	no := &corev1.Node{}
	err := r.Get(ctx, o.NamespacedName, no)
	if err != nil {
		return reconcile.Result{}, err
	}

	var labelKey string
	if r.opts.UseAwsLifecycle {
		// 0. useless check for label
		// 1. get the instance id (and check if in cache)
		parts := strings.Split(no.Spec.ProviderID, "/")
		id := parts[len(parts)-1]
		// 2. get the aws instance details (better from a cache)
		roleSuffix, err := r.cache.Get(id)
		if err != nil {
			return reconcile.Result{}, err
		}
		labelKey = LabelPrefix + roleSuffix

		no.ObjectMeta.Labels[labelKey] = ""
	}

	if r.opts.FromLabel != "" {
		if v, ok := no.ObjectMeta.Labels[r.opts.FromLabel]; !ok {
			return reconcile.Result{}, LabelSourceUndefinedErr
		} else {
			labelKey = LabelPrefix +
				r.opts.FromLabelPrefix + v +
				r.opts.FromLabelSuffix
			no.ObjectMeta.Labels[labelKey] = ""
		}
	}

	if labelKey == "" {
		return reconcile.Result{}, LabelSourceUndefinedErr
	}

	if err := r.Update(ctx, no); err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
