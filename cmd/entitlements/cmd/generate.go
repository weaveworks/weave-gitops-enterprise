package cmd

import (
	"crypto/ed25519"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/cobra"
	"github.com/weaveworks/libgitops/pkg/serializer"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubectl/pkg/scheme"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate an entitlement for WeGO EE",
	PreRun: func(cmd *cobra.Command, args []string) {
		if durationYears <= 0 && durationMonths <= 0 && durationDays <= 0 {
			log.Fatal("The duration of the entitlements has not been set. Please use the flags '-y', '-m' or '-d' to specify a duration.")
		}
	},
	Run: func(cmd *cobra.Command, args []string) {

		privatePem, err := ioutil.ReadFile(privateKeyFilename)
		if err != nil {
			log.Fatal(err)
		}
		privateKey, err := jwt.ParseEdPrivateKeyFromPEM(privatePem)
		if err != nil {
			log.Fatal(err)
		}
		pk := privateKey.(ed25519.PrivateKey)

		now := time.Now().UTC()
		expiresAt := now.AddDate(durationYears, durationMonths, durationDays)
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
			log.Fatal(err)
		}

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: "wego-ee-entitlement",
			},
			Type:       "Opaque",
			StringData: map[string]string{"entitlement": tokenString},
		}
		s := serializer.NewSerializer(scheme.Scheme, nil)
		fw := serializer.NewYAMLFrameWriter(os.Stdout)
		s.Encoder().Encode(fw, secret)
	},
}

var (
	privateKeyFilename string
	customerEmail      string
	durationYears      int
	durationMonths     int
	durationDays       int
)

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.PersistentFlags().StringVarP(&privateKeyFilename, "private-key-filename", "p", "", "")
	generateCmd.MarkPersistentFlagRequired("private-key-filename")
	generateCmd.PersistentFlags().StringVarP(&customerEmail, "customer-email", "c", "", "")
	generateCmd.MarkPersistentFlagRequired("customer-email")
	generateCmd.PersistentFlags().IntVarP(&durationYears, "years", "y", 0, "")
	generateCmd.PersistentFlags().IntVarP(&durationMonths, "months", "m", 0, "")
	generateCmd.PersistentFlags().IntVarP(&durationDays, "days", "d", 0, "")
}
