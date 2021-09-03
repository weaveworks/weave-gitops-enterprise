package cmd

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("generate called")

		pub, priv, err := ed25519.GenerateKey(nil)
		fmt.Println(hex.EncodeToString(pub))
		fmt.Println(hex.EncodeToString(priv))
		fmt.Println(err)

		claims := jwt.StandardClaims{
			Issuer:    "sales@weave.works",
			NotBefore: time.Now().UTC().Unix(),
			ExpiresAt: time.Now().UTC().AddDate(1, 0, 0).Unix(),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
		tokenString, err := token.SignedString(priv)

		fmt.Println(tokenString, err)
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
