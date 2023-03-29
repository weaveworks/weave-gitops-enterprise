package kubefakes

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"net/http"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

type fakeControllerManager struct {
	log    logr.Logger
	scheme *runtime.Scheme
}

func (f fakeControllerManager) SetFields(i interface{}) error {
	f.log.Info("faked")
	return nil
}

func (f fakeControllerManager) GetConfig() *rest.Config {
	f.log.Info("faked")
	return nil
}

func (f fakeControllerManager) GetScheme() *runtime.Scheme {
	f.log.Info("faked get scheme")
	return f.scheme
}

func (f fakeControllerManager) GetClient() client.Client {
	s := runtime.NewScheme()
	return fake.NewClientBuilder().WithScheme(s).Build()
}

func (f fakeControllerManager) GetFieldIndexer() client.FieldIndexer {
	f.log.Info("faked")
	return nil
}

func (f fakeControllerManager) GetCache() cache.Cache {
	f.log.Info("faked")
	return nil
}

func (f fakeControllerManager) GetEventRecorderFor(name string) record.EventRecorder {
	f.log.Info("faked")
	return nil
}

func (f fakeControllerManager) GetRESTMapper() meta.RESTMapper {
	f.log.Info("faked")
	return nil
}

func (f fakeControllerManager) GetAPIReader() client.Reader {
	f.log.Info("faked")
	return nil
}

func (f fakeControllerManager) Start(ctx context.Context) error {
	f.log.Info("faked")
	return nil
}

func (f fakeControllerManager) Add(runnable manager.Runnable) error {
	f.log.Info("faked")
	return nil
}

func (f fakeControllerManager) Elected() <-chan struct{} {
	f.log.Info("faked")
	return nil
}

func (f fakeControllerManager) AddMetricsExtraHandler(path string, handler http.Handler) error {
	f.log.Info("faked")
	return nil
}

func (f fakeControllerManager) AddHealthzCheck(name string, check healthz.Checker) error {
	//TODO implement me
	panic("implement me")
}

func (f fakeControllerManager) AddReadyzCheck(name string, check healthz.Checker) error {
	//TODO implement me
	panic("implement me")
}

func (f fakeControllerManager) GetWebhookServer() *webhook.Server {
	f.log.Info("faked")
	return nil
}

func (f fakeControllerManager) GetLogger() logr.Logger {
	f.log.Info("faked")
	return f.log
}

func (f fakeControllerManager) GetControllerOptions() v1alpha1.ControllerConfigurationSpec {
	f.log.Info("faked")
	return v1alpha1.ControllerConfigurationSpec{}
}

func NewControllerManager(config *rest.Config, options ctrl.Options) (ctrl.Manager, error) {
	options.Logger.Info("created fake watcher")

	return fakeControllerManager{log: options.Logger, scheme: options.Scheme}, nil
}
