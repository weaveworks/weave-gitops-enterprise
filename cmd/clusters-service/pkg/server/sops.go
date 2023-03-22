package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	gopgp "github.com/ProtonMail/gopenpgp/v2/crypto"
	kustomizev1beta2 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	capiv1_proto "github.com/weaveworks/weave-gitops-enterprise/cmd/clusters-service/pkg/protos"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"go.mozilla.org/sops/v3"
	"go.mozilla.org/sops/v3/aes"
	"go.mozilla.org/sops/v3/age"
	"go.mozilla.org/sops/v3/cmd/sops/common"
	"go.mozilla.org/sops/v3/cmd/sops/formats"
	"go.mozilla.org/sops/v3/keys"
	pgp "go.mozilla.org/sops/v3/pgp"
	structpb "google.golang.org/protobuf/types/known/structpb"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	EncryptedRegex                   = "^(data|stringData)$"
	DecryptionPGPExt                 = ".asc"
	DecryptionAgeExt                 = ".agekey"
	SopsPublicKeyNameAnnotation      = "sops-public-key/name"
	SopsPublicKeyNamespaceAnnotation = "sops-public-key/namespace"
)

type Encryptor struct {
	store common.Store
}

func NewEncryptor() *Encryptor {
	return &Encryptor{
		store: common.StoreForFormat(formats.Json),
	}
}

func (e *Encryptor) Encrypt(raw []byte, keys ...keys.MasterKey) ([]byte, error) {
	branches, err := e.store.LoadPlainFile(raw)
	if err != nil {
		return nil, err
	}

	tree := sops.Tree{
		Branches: branches,
		Metadata: sops.Metadata{
			EncryptedRegex: EncryptedRegex,
			KeyGroups: []sops.KeyGroup{
				keys,
			},
		},
	}

	dataKey, errs := tree.GenerateDataKey()
	if errs != nil {
		return nil, fmt.Errorf("failed to get data key: %v", errs)
	}

	err = common.EncryptTree(common.EncryptTreeOpts{
		Cipher:  aes.NewCipher(),
		DataKey: dataKey,
		Tree:    &tree,
	})

	if err != nil {
		return nil, err
	}

	encrypted, err := e.store.EmitEncryptedFile(tree)
	if err != nil {
		return nil, err
	}

	return encrypted, nil
}

func (e *Encryptor) EncryptWithPGP(raw []byte, publicKey string) ([]byte, error) {
	err := importPGPKey(publicKey)
	if err != nil {
		return nil, err
	}

	key, err := gopgp.NewKeyFromArmoredReader(strings.NewReader(publicKey))
	if err != nil {
		return nil, err
	}

	masterKey := pgp.NewMasterKeyFromFingerprint(key.GetFingerprint())
	return e.Encrypt(raw, masterKey)
}

func (e *Encryptor) EncryptWithAGE(raw []byte, recipient string) ([]byte, error) {
	keys, err := age.MasterKeysFromRecipients(recipient)
	if err != nil {
		return nil, fmt.Errorf("failed to create the master key: %w", err)
	}

	if keys == nil {
		return nil, errors.New("no key found")
	}

	return e.Encrypt(raw, keys[0])
}

func (s *server) EncryptSopsSecret(ctx context.Context, msg *capiv1_proto.EncryptSopsSecretRequest) (*capiv1_proto.EncryptSopsSecretResponse, error) {
	clustersClient, err := s.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	var kustomization kustomizev1beta2.Kustomization
	kustomizationKey := client.ObjectKey{
		Name:      msg.KustomizationName,
		Namespace: msg.KustomizationNamespace,
	}
	if err := clustersClient.Get(ctx, msg.ClusterName, kustomizationKey, &kustomization); err != nil {
		return nil, fmt.Errorf("failed to get kustomization: %w", err)
	}

	if kustomization.Spec.Decryption == nil {
		return nil, errors.New("kustomization missing decryption settings")
	}

	if kustomization.Spec.Decryption.SecretRef == nil {
		return nil, errors.New("kustomization doesn't have decryption secret")
	}

	encryptionSecretName, ok := kustomization.Annotations[SopsPublicKeyNameAnnotation]
	if !ok {
		return nil, errors.New("kustomization is missing encryption key information")
	}

	encryptionSecretNamespace, ok := kustomization.Annotations[SopsPublicKeyNamespaceAnnotation]
	if !ok {
		return nil, errors.New("kustomization is missing encryption key information")
	}

	encryptionSecret := client.ObjectKey{
		Name:      encryptionSecretName,
		Namespace: encryptionSecretNamespace,
	}

	var decryptionSecret v1.Secret
	if err := clustersClient.Get(ctx, msg.ClusterName, encryptionSecret, &decryptionSecret); err != nil {
		return nil, fmt.Errorf("failed to get encryption key: %w", err)
	}

	rawSecret, err := generateSecret(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secret: %w", err)
	}

	encryptedKey, err := encryptSecret(rawSecret, decryptionSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt secret: %w", err)
	}

	result := structpb.Value{}
	err = result.UnmarshalJSON(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal secret: %w", err)
	}

	return &capiv1_proto.EncryptSopsSecretResponse{
		EncryptedSecret: &result,
		Path:            kustomization.Spec.Path,
	}, nil
}

func generateSecret(msg *capiv1_proto.EncryptSopsSecretRequest) ([]byte, error) {
	data := map[string][]byte{}
	for key := range msg.Data {
		value := msg.Data[key]
		data[key] = []byte(value)
	}

	secret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1.SchemeGroupVersion.Identifier(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      msg.Name,
			Namespace: msg.Namespace,
			Labels:    msg.Labels,
		},
		Data:       data,
		StringData: msg.StringData,
		Immutable:  &msg.Immutable,
		Type:       v1.SecretType(msg.Type),
	}

	var buf bytes.Buffer
	serializer := json.NewSerializer(json.DefaultMetaFactory, nil, nil, false)
	if err := serializer.Encode(&secret, &buf); err != nil {
		return nil, fmt.Errorf("failed to serialize object: %w", err)
	}

	return buf.Bytes(), nil
}

func encryptSecret(raw []byte, decryptSecret v1.Secret) ([]byte, error) {
	encryptor := NewEncryptor()
	for name, value := range decryptSecret.Data {
		switch filepath.Ext(name) {
		case DecryptionPGPExt:
			return encryptor.EncryptWithPGP(raw, string(value))
		case DecryptionAgeExt:
			return encryptor.EncryptWithAGE(raw, strings.TrimRight(string(value), "\n"))
		default:
			return nil, errors.New("invalid decryption secret")
		}
	}
	return nil, nil
}

func importPGPKey(pk string) error {
	binary := "gpg"
	if envBinary := os.Getenv("SOPS_GPG_EXEC"); envBinary != "" {
		binary = envBinary
	}
	args := []string{"--batch", "--import"}
	cmd := exec.Command(binary, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdin = bytes.NewReader([]byte(pk))
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	return cmd.Run()
}

func (s *server) ListSopsKustomizations(ctx context.Context, req *capiv1_proto.ListSopsKustomizationsRequest) (*capiv1_proto.ListSopsKustomizationsResponse, error) {

	clustersClient, err := s.clustersManager.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), req.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	if clustersClient == nil {
		return nil, fmt.Errorf("cluster %s not found", req.ClusterName)
	}

	kustomizations := []*capiv1_proto.SopsKustomizations{}
	kustomizationList := &kustomizev1beta2.KustomizationList{}

	if err := clustersClient.List(ctx, req.ClusterName, kustomizationList); err != nil {
		return nil, fmt.Errorf("failed to list kustomizations, error: %w", err)
	}

	for _, kustomization := range kustomizationList.Items {
		if kustomization.Spec.Decryption != nil && strings.EqualFold(kustomization.Spec.Decryption.Provider, "sops") {
			if kustomization.Annotations[SopsPublicKeyNameAnnotation] != "" && kustomization.Annotations[SopsPublicKeyNamespaceAnnotation] != "" {
				kustomizations = append(kustomizations, &capiv1_proto.SopsKustomizations{
					Name:      kustomization.Name,
					Namespace: kustomization.Namespace,
				})
			}
		}
	}

	response := capiv1_proto.ListSopsKustomizationsResponse{
		Kustomizations: kustomizations,
		Total:          int32(len(kustomizations)),
	}

	return &response, nil
}
