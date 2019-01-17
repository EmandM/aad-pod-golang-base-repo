package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/auth"
)

func hello(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Emmcmill Keyvault Sample")
}

func getKeyvaultSecret(w http.ResponseWriter, r *http.Request) {
	keyvaultName := os.Getenv("AZURE_KEYVAULT_NAME")
	keyvaultSecretName := os.Getenv("AZURE_KEYVAULT_SECRET_NAME")
	keyvaultSecretVersion := os.Getenv("AZURE_KEYVAULT_SECRET_VERSION")

	keyClient := keyvault.New()
	authorizer, err := auth.NewAuthorizerFromEnvironment()

	if err == nil {
		keyClient.Authorizer = authorizer
	} else {
		io.WriteString(w, fmt.Sprintf("Failed to create authorizer: %v", err))
		log.Printf("Failed to create authorizer: %v", err)
		return
	}

	secret, err := keyClient.GetSecret(context.Background(), fmt.Sprintf("https://%s.vault.azure.net", keyvaultName), keyvaultSecretName, keyvaultSecretVersion)
	if err != nil {
		io.WriteString(w, fmt.Sprintf("Failed to retrieve the Keyvault secret: %v", err))
		log.Printf("failed to retrieve the Keyvault secret: %v", err)
		return
	}

	io.WriteString(w, fmt.Sprintf("The value of the Keyvault secret is: %v", *secret.Value))
}

func main() {
	var PORT string
	if PORT = os.Getenv("PORT"); PORT == "" {
		PORT = "8080"
	}
	http.HandleFunc("/", hello)
	http.HandleFunc("/keyvault", getKeyvaultSecret)
	log.Println("http server listening on :"+PORT)
	http.ListenAndServe(":"+PORT, nil)
}
