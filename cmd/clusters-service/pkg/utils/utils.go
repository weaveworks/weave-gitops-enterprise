package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"

	capiv1 "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/api/v1alpha1"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/yaml"
)

// Signature for json.MarshalIndent accepted as a method parameter for unit tests
type MarshalIndent func(v interface{}, prefix, indent string) ([]byte, error)

func GetClientset() (*kubernetes.Clientset, *rest.RESTClient, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("error creating in-cluster config: %s", err)
	}

	crdRestClient, err := getRestClientForConfig(config)
	if err != nil {
		panic(err)
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating kubernetes clientset: %s", err)
	}

	return clientset, crdRestClient, nil
}

func GetClientsetFromKubeconfig() (*kubernetes.Clientset, *rest.RESTClient, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}

	crdRestClient, err := getRestClientForConfig(config)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientset, crdRestClient, nil
}

func getRestClientForConfig(config *rest.Config) (*rest.RESTClient, error) {
	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &capiv1.GroupVersion
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	crdRestClient, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		panic(err)
	}
	return crdRestClient, nil
}

type errorView struct {
	Error string `json:"message"`
}

func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}, marshalIndentFn MarshalIndent) {
	response, err := marshalIndentFn(payload, "", " ")
	if err != nil {
		WriteError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = w.Write(response)
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), 500)
	}
}

// WriteError writes the error into the request
func WriteError(w http.ResponseWriter, appError error, code int) {
	log.Error(appError)
	payload, err := json.Marshal(errorView{Error: appError.Error()})
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), 500)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintln(w, string(payload))
}

func B64ResourceTemplate(ct []capiv1.ResourceTemplate) ([]byte, error) {
	b, err := yaml.Marshal(ct)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal resource template: %s", err)
	}

	rtBytesB64 := Base64Encode(b)
	return rtBytesB64, nil
}

func Base64Encode(message []byte) []byte {
	b := make([]byte, base64.StdEncoding.EncodedLen(len(message)))
	base64.StdEncoding.Encode(b, message)
	return b
}

func Base64Decode(message []byte) (b []byte, err error) {
	var l int
	b = make([]byte, base64.StdEncoding.DecodedLen(len(message)))
	l, err = base64.StdEncoding.Decode(b, message)
	if err != nil {
		return
	}
	return b[:l], nil
}
