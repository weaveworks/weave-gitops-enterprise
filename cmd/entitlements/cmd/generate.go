package cmd

import (
	"log"
	"os"

	"github.com/weaveworks/weave-gitops-enterprise/pkg/entitlement"

	"github.com/spf13/cobra"
	"github.com/weaveworks/libgitops/pkg/serializer"
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
		token, err := entitlement.CreateEntitlement(privateKeyFilename, durationYears, durationMonths, durationDays, customerEmail)
		cobra.CheckErr(err)
		secret := entitlement.WrapEntitlementInSecret(token, secretName, secretNamespace)
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
	secretName         string
	secretNamespace    string
)

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.LocalFlags().StringVarP(&privateKeyFilename, "private-key-filename", "p", "", "The private key to use for signing the entitlement")
	generateCmd.MarkFlagRequired("private-key-filename")
	generateCmd.LocalFlags().StringVarP(&customerEmail, "customer-email", "c", "", "The email of the customer that the entitlement is generated for")
	generateCmd.MarkFlagRequired("customer-email")
	generateCmd.LocalFlags().IntVarP(&durationYears, "duration-years", "y", 0, "Number of years to use for the duration of the entitlement")
	generateCmd.LocalFlags().IntVarP(&durationMonths, "duration-months", "m", 0, "Number of months to use for the duration of the entitlement")
	generateCmd.LocalFlags().IntVarP(&durationDays, "duration-days", "d", 0, "Number of days to use for the duration of the entitlement")
	generateCmd.LocalFlags().StringVarP(&secretName, "name", "n", "wego-ee-entitlement", "The name of the secret")
	generateCmd.LocalFlags().StringVar(&secretNamespace, "namespace", "default", "The namespace of the secret")
}
