package simulator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"3Xpl0it3r.com/kube-simulator/pkg/agent"
	"3Xpl0it3r.com/kube-simulator/pkg/apiclient"
	"3Xpl0it3r.com/kube-simulator/pkg/cluster"
	"3Xpl0it3r.com/kube-simulator/pkg/util"
	myutil "3Xpl0it3r.com/kube-simulator/pkg/util"

	kvapp "github.com/k3s-io/kine/pkg/app"
	kvep "github.com/k3s-io/kine/pkg/endpoint"
	"github.com/pkg/errors"
)

const (
	ClusterDefaultName             = "kuberntetes"
	KubeGroupWithDefault           = "kubernetes"
	KubeGroupWithAdmin             = "system:masters"
	KubeGroupWithControllerManager = "system:kube-controller-manager"
	KubeGroupWithScheduler         = "system:kube-scheduler"
	KubeConfigControllerManager    = "controller-manager.yml"
	KubeConfigScheduler            = "scheduler.yml"
	KubeConfigAdmin                = "kube-admin.yml"
	KubeUserAdmin                  = "kuberntetes-admin"
)

var (
	kvStorageLogger = util.NewLogger("kv")
	bootstraplogger = util.NewLogger("bootstrap")
)

// Start [#TODO](should add some comments)
func Start(parent context.Context, config Config) error {
	// prepare some necessary certificated file for all k8s components
	if err := prepareAllNecessaryClusterCertificates(&config); err != nil {
		return errors.Wrap(err, "bootstrap certificated failed")
	}
	// prepare kubeconfig for some clients like kube-controller/scheduler/kubelet.... to access apiserver
	if err := prepareComponentClusterConfigs(&config.Cluster); err != nil {
		return errors.Wrap(err, "bootstrap some kubeconfigs failed")
	}
	// run kv storage(mock etcd) and wait kv storage ready then go on
	runKvStorage(&config.Etcd)
	if err := waitForKvStorageReady("", ""); err != nil {
		return err
	}

	if err := cluster.Run(&config.Cluster, parent.Done()); err != nil {
		return errors.Wrap(err, "start cluster failed")
	}

	if err := agent.Run(&config.Agent); err != nil {
		return errors.Wrap(err, "start agent failed")
	}

	client, err := apiclient.NewClusterClient("", config.Cluster.ClientConfigFile.Administrator)
	if err != nil {
		return errors.Wrap(err, "unable to build client throug kubeconfig, cluster may fail into trouble")
	}

	if err := bootstrapKubeadmCompatibleRBAC(client); err != nil {
		bootstraplogger.Error("bootstrap rabc failed")
	}
	if err := bootstrapOptionalPresetNodes(config.NodeNum, client); err != nil {
		bootstraplogger.Error("bootstrap nodes failed")
	}

	return nil

}
func runKvStorage(etcd *EtcdConfig) {
	argsMap := map[string]string{
		"ca-file":          etcd.CACert.CertFile,
		"server-cert-file": etcd.ServerCert.CertFile,
		"server-key-file":  etcd.ServerCert.KeyFile,
	}
	args := myutil.GetArgsList(argsMap, nil)
	config := kvapp.Config(args)
	config.Listener = etcd.Listener
	config.WaitGroup = &sync.WaitGroup{}
	config.Endpoint = fmt.Sprintf("sqlite://%s/simukube.db?mode=rwc&_journal_mode=WAL", etcd.DataDir)
	kvStorageLogger.Infof("etcd datadir is %s", config.Endpoint)
	go func() {
		kvStorageLogger.Infof("Running kv-storage")
		_, err := kvep.Listen(context.Background(), config)
		config.WaitGroup.Wait()
		kvStorageLogger.Errorf("kv storage existed: %v", err)
	}()
}

func waitForKvStorageReady(caCertPath, address string) error {
	// todo
	time.Sleep(5 * time.Second)
	return nil
}
