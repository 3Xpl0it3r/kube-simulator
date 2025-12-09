package cluster

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	cliflag "k8s.io/component-base/cli/flag"
	logsapi "k8s.io/component-base/logs/api/v1"
	"k8s.io/component-base/version/verflag"
	schedulerapp "k8s.io/kubernetes/cmd/kube-scheduler/app"
	schedleroptions "k8s.io/kubernetes/cmd/kube-scheduler/app/options"
)

func NewRewriteSchedulerCommand() *cobra.Command {
	opts := schedleroptions.NewOptions()

	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			return runModifiedSchedulerCommand(cmd, opts)
		},
	}
	nfs := opts.Flags
	fs := cmd.Flags()
	for _, f := range nfs.FlagSets {
		fs.AddFlagSet(f)
	}

	return cmd
}

// runModifiedSchedulerCommand runs the scheduler.
func runModifiedSchedulerCommand(cmd *cobra.Command, opts *schedleroptions.Options) error {
	verflag.PrintAndExitIfRequested()

	// Activate logging as soon as possible, after that
	// show flags with the final logging configuration.
	if err := logsapi.ValidateAndApply(opts.Logs, utilfeature.DefaultFeatureGate); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	cliflag.PrintFlags(cmd.Flags())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc, sched, err := schedulerapp.Setup(ctx, opts)
	if err != nil {
		return err
	}
	// add feature enablement metrics
	utilfeature.DefaultMutableFeatureGate.AddMetrics()
	if err := schedulerapp.Run(ctx, cc, sched); err != nil {
		return err
	}
	<-ctx.Done()
	return ctx.Err()
}
