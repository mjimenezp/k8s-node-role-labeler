package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2/klogr"

	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	"github.com/safanaj/k8s-node-role-labeler/pkg/awsutils"
	"github.com/safanaj/k8s-node-role-labeler/pkg/reconcilers"
)

var version, progname string

func main() {
	flags := parseFlags()
	if flags.version {
		fmt.Printf("%s version %s\n", progname, version)
		os.Exit(0)
	}

	reconcilers.LabelPrefix = flags.labelPrefix

	ctx, ctxCancel := context.WithCancel(context.Background())
	log := klogr.New().WithName(progname)
	logf.SetLogger(log)

	cfg := config.GetConfigOrDie()
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		log.Error(err, "creating manager")
		os.Exit(1)
	}

	awsCache := awsutils.NewCache(ctx, flags.tagKey)
	inject.LoggerInto(awsCache, log.WithName("awsCache"))
	awsCache.Start()

	err = builder.
		ControllerManagedBy(mgr).
		For(&corev1.Node{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				for l, _ := range e.Object.GetLabels() {
					if strings.HasPrefix(l, reconcilers.LabelPrefix) {
						return false
					}
				}
				return true
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				for l, _ := range e.ObjectNew.GetLabels() {
					if strings.HasPrefix(l, reconcilers.LabelPrefix) {
						return false
					}
				}
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				no, ok := e.Object.(*corev1.Node)
				if ok {
					parts := strings.Split(no.Spec.ProviderID, "/")
					id := parts[len(parts)-1]
					awsCache.Del(id)
				}
				return false
			},
		}).
		Complete(reconcilers.NewNodeRoleLabelReconciler(awsCache))

	if err := mgr.Start(ctx); err != nil {
		log.Error(err, "unable to run manager")
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	close(quit)
	ctxCancel()
	os.Exit(0)
}
