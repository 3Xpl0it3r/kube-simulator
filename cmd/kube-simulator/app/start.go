package app

import (
	"context"
	"fmt"
	"os"

	"3Xpl0it3r.com/kube-simulator/cmd/kube-simulator/options"
	"3Xpl0it3r.com/kube-simulator/pkg/simulator"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	logsapi "k8s.io/component-base/logs/api/v1"
)

func init() {
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)
	logsapi.ReapplyHandling = logsapi.ReapplyHandlingIgnoreUnchanged
}

func NewKubeSimulatorCommand() *cobra.Command {
	opts := options.NewOptions()
	cmd := &cobra.Command{
		Short: "Launch kube-simulator",
		Long:  "Launch kube-simulator",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.Validate(); err != nil {
				return fmt.Errorf("Options validate failed %v. ", err)
			}
			if err := opts.Complete(); err != nil {
				return fmt.Errorf("Options Complete failed %v. ", err)
			}
			if err := RunCommand(opts); err != nil {
				return fmt.Errorf("Run %s failed. Err: %v", os.Args[0], err)
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return preRunE(opts)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	fs := cmd.Flags()
	fs.AddFlagSet(opts.FlagsSets())

	return cmd
}

func RunCommand(o *options.Options) error {
	ctx := SetupSignalHandler(context.Background())
	config := o.Config()
	if err := config.Complete(); err != nil {
		return err
	}
	if err := simulator.Start(ctx, config); err != nil {
		return err
	}
	<-ctx.Done()
	return ctx.Err()
}

func preRunE(o *options.Options) error {
	if o.ResetCluster {
		os.RemoveAll(o.DataDir)
	}
	if err := ensureDir(o.Simulator.Etcd.DataDir); err != nil {
		return err
	}

	return nil
}

func ensureDir(dirPath string) error {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("create data dir %s failed %s", dirPath, err)
	}
	return nil
}
