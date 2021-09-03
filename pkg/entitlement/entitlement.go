package entitlement

import (
	"crypto/ed25519"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/golang-jwt/jwt/v4"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateEntitlement(privateKeyFilename string, years, months, days int, customerEmail string) (string, error) {
	privatePem, err := ioutil.ReadFile(privateKeyFilename)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	privateKey, err := jwt.ParseEdPrivateKeyFromPEM(privatePem)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	pk := privateKey.(ed25519.PrivateKey)

	now := time.Now().UTC()
	expiresAt := now.AddDate(years, months, days)
	claims := jwt.StandardClaims{
		Issuer:    "sales@weave.works",
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		ExpiresAt: expiresAt.Unix(),
		Subject:   customerEmail,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	tokenString, err := token.SignedString(pk)
	if err != nil {
		return "", fmt.Errorf("%w", err)
	}
	return tokenString, nil
}

func WrapEntitlementInSecret(entitlement, name, namespace string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Type:       "Opaque",
		StringData: map[string]string{"entitlement": entitlement},
	}
}
