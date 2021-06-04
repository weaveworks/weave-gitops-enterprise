package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/tools/clientcmd"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Signature for json.MarshalIndent accepted as a method parameter for unit tests
type MarshalIndent func(v interface{}, prefix, indent string) ([]byte, error)

func GetClientset() (*kubernetes.Clientset, error) {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("error creating in-cluster config: %s", err)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes clientset: %s", err)
	}

	return clientset, nil
}

func GetClientsetFromKubeconfig() (*kubernetes.Clientset, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientset, nil
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
	w.Write(response)
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
