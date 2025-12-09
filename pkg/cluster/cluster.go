package cluster

import (
	"context"
	"net/http"
	"time"

	kuberesource "3Xpl0it3r.com/kube-simulator/pkg/kuberes"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	apiserverapp "k8s.io/kubernetes/cmd/kube-apiserver/app"
	controllermgapp "k8s.io/kubernetes/cmd/kube-controller-manager/app"
)

var (
	loggerForApiServer    = logrus.WithField("component", "apiserver")
	loggerForScheduler    = logrus.WithField("component", "scheduler")
	loggerForControllerMg = logrus.WithField("component", "controller-manager")
)

func Run(config *Config) error {
	// run apiserver async
	runApiServer(config)
	if err := waitForApiServerRunning(config); err != nil {
		return errors.Wrap(err, "wait apiserver ready failed")
	}
	runControllerManager(config)
	// run kube-scheduler
	runScheduler(config)
	return nil
}

// sync run apiserver
func runApiServer(config *Config) {
	argsMap := map[string]string{
		"secure-port":                      config.ListenPort,
		"advertise-address":                config.ListenHost,
		"bind-address":                     config.ListenHost,
		"service-cluster-ip-range":         config.ServiceCIDR,
		"api-audiences":                    "https://kubernetes.default.svc.cluster.local",
		"client-ca-file":                   config.TLS.CA.CertFile,
		"tls-cert-file":                    config.TLS.Server.CertFile,
		"tls-private-key-file":             config.TLS.Server.KeyFile,
		"service-account-issuer":           "https://kubernetes.default.svc.cluster.local",
		"service-account-key-file":         config.TLS.ServiceAccountKeyFile,
		"service-account-signing-key-file": config.TLS.ServiceAccountSigningKeyFile,
		"authorization-mode":               config.AuthorizationMode,
		"etcd-cafile":                      config.TLS.EtcdCA,
		"etcd-certfile":                    config.TLS.EtcdClient.CertFile,
		"etcd-keyfile":                     config.TLS.EtcdClient.KeyFile,
		"etcd-servers":                     "127.0.0.1:2379",
	}

	args := GetArgsList(argsMap, nil)

	command := apiserverapp.NewAPIServerCommand()
	command.SetArgs(args)

	go func() {
		loggerForApiServer.Infof("Running kube-apiserver %s", args)
		loggerForApiServer.Fatalf("apiserver existed %v", command.Execute())
	}()
}

func runScheduler(config *Config) {
	argsMap := map[string]string{
		"kubeconfig":                config.ClientConfigFile.Scheduler,
		"authentication-kubeconfig": config.ClientConfigFile.Scheduler,
		"authorization-kubeconfig":  config.ClientConfigFile.Scheduler,
	}
	args := GetArgsList(argsMap, nil)
	command := NewRewriteSchedulerCommand()
	command.SetArgs(args)
	go func() {
		loggerForScheduler.Infof("Running kube-scheduler %s", args)
		loggerForScheduler.Fatalf("kube-scheduler existed %v", command.Execute())
	}()

}

func runControllerManager(config *Config) {
	argsMap := map[string]string{
		"kubeconfig":                       config.ClientConfigFile.ControllerManager,
		"authentication-kubeconfig":        config.ClientConfigFile.ControllerManager,
		"authorization-kubeconfig":         config.ClientConfigFile.ControllerManager,
		"root-ca-file":                     config.TLS.CA.CertFile,
		"cluster-signing-cert-file":        config.TLS.CA.CertFile,
		"cluster-signing-key-file":         config.TLS.CA.KeyFile,
		"service-account-private-key-file": config.TLS.ServiceAccountSigningKeyFile,
		"cluster-cidr":                     config.ClusterCIDR,
		"controllers":                      "*,bootstrapsigner,tokencleaner",
		"allocate-node-cidrs":              "false",
		"use-service-account-credentials":  "true",
	}

	args := GetArgsList(argsMap, nil)

	command := controllermgapp.NewControllerManagerCommand()
	command.SetArgs(args)

	go func() {
		loggerForControllerMg.Infof("Running kube-controller-manager %s", args)
		loggerForControllerMg.Fatalf("controller manager existed: %v", command.Execute())
	}()
}

func waitForApiServerRunning(config *Config) error {
	client, err := kuberesource.NewClusterClient("", config.ClientConfigFile.Administrator)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	healthStatus := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			result := client.Discovery().RESTClient().Get().AbsPath("/healthz").Do(context.Background()).StatusCode(&healthStatus)
			if result.Error() != nil || healthStatus != http.StatusOK {
				time.Sleep(10 * time.Second)
				loggerForApiServer.Debug("waiting apiserver ready....")
				continue
			}
			return nil
		}
	}
}
