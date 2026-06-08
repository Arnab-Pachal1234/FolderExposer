package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var keygenCmd = &cobra.Command{
	Use:   "keygen",
	Short: "Generate a new 4096-bit RSA encryption key for the tunnel",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("⏳ Generating 4096-bit RSA key... This requires heavy computation, please wait.")

		// 1. Generate the math for the key
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			fmt.Printf("❌ Error generating key: %v\n", err)
			return
		}

		// 2. Safely create the certs directory (0700 means only the owner can open this folder)
		err = os.MkdirAll("certs", 0700)
		if err != nil {
			fmt.Printf("❌ Error creating certs directory: %v\n", err)
			return
		}

		// 3. Convert the raw math into a PEM format (Standard encryption format)
		privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
		privateKeyPEM := &pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privateKeyBytes,
		}

		// 4. Create and lock the file
		keyPath := filepath.Join("certs", "tunnel.key")
		keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) // 0600 = strict read/write permissions
		if err != nil {
			fmt.Printf("❌ Error saving key file: %v\n", err)
			return
		}
		defer keyFile.Close()

		// 5. Write the key to the file
		pem.Encode(keyFile, privateKeyPEM)

		fmt.Println("\n=======================================================")
		fmt.Println("🔐 SUCCESS! ENCRYPTION KEY GENERATED")
		fmt.Println("=======================================================")
		fmt.Printf("Path: %s\n", keyPath)
		fmt.Println("Your tunnel traffic is now cryptographically secure.")
		fmt.Println("=======================================================")
	},
}

func init() {
	// Attach this new command to your main CLI
	rootCmd.AddCommand(keygenCmd)
}
