package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/Irongoshan-ux/gophkeeper/internal/client"
	"github.com/Irongoshan-ux/gophkeeper/internal/client/storage"
	"github.com/Irongoshan-ux/gophkeeper/internal/config"
	"github.com/Irongoshan-ux/gophkeeper/internal/model"
	"github.com/Irongoshan-ux/gophkeeper/internal/validation"
	"github.com/Irongoshan-ux/gophkeeper/pkg/crypto"
	"github.com/spf13/cobra"
)

func syncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Synchronize secrets with the server",
		Run: func(_ *cobra.Command, _ []string) {
			api, cfg, err := connectFromConfig()
			if err != nil {
				die(err)
			}
			defer api.Close()
			if err := runSync(api, cfg); err != nil {
				die(err)
			}
			fmt.Println("Sync complete.")
		},
	}
}

func runSync(api *client.API, cfg *config.ClientConfig) error {
	cache, err := storage.LoadCache()
	if err != nil {
		return err
	}
	remote, err := api.Sync(context.Background(), cache.LastSync)
	if err != nil {
		return err
	}
	cache.Secrets = storage.MergeSecrets(cache.Secrets, remote)
	cache.LastSync = time.Now().UTC()
	return storage.SaveCache(cache)
}

func newSecretCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secret",
		Short: "Manage secrets",
	}
	cmd.AddCommand(secretCreateCmd(), secretGetCmd(), secretListCmd(), secretDeleteCmd())
	return cmd
}

func secretCreateCmd() *cobra.Command {
	var name, secretType, login, pass, text, metadata, cardNumber string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new secret",
		Run: func(_ *cobra.Command, _ []string) {
			api, cfg, err := connectFromConfig()
			if err != nil {
				die(err)
			}
			defer api.Close()

			key, err := deriveKeyFromConfig(cfg)
			if err != nil {
				die(err)
			}

			st, payload, err := BuildPayload(secretType, login, pass, text, metadata, cardNumber)
			if err != nil {
				die(err)
			}
			encrypted, err := crypto.Encrypt(key, payload)
			if err != nil {
				die(err)
			}

			secret, err := api.CreateSecret(context.Background(), st, name, encrypted)
			if err != nil {
				die(err)
			}
			fmt.Printf("Created secret %s (%s)\n", secret.ID, secret.Name)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "secret name")
	cmd.Flags().StringVar(&secretType, "type", "credentials", "secret type: credentials|text|binary|card|otp")
	cmd.Flags().StringVar(&login, "login", "", "login for credentials")
	cmd.Flags().StringVar(&pass, "password", "", "password for credentials")
	cmd.Flags().StringVar(&text, "text", "", "text content")
	cmd.Flags().StringVar(&metadata, "metadata", "", "metadata label")
	cmd.Flags().StringVar(&cardNumber, "card", "", "card number")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func secretGetCmd() *cobra.Command {
	var id string
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get and decrypt a secret",
		Run: func(_ *cobra.Command, _ []string) {
			api, cfg, err := connectFromConfig()
			if err != nil {
				die(err)
			}
			defer api.Close()

			key, err := deriveKeyFromConfig(cfg)
			if err != nil {
				die(err)
			}

			secret, err := api.GetSecret(context.Background(), id)
			if err != nil {
				die(err)
			}
			payload, err := crypto.Decrypt(key, secret.EncryptedData)
			if err != nil {
				die(err)
			}
			fmt.Printf("ID: %s\nName: %s\nType: %d\nMetadata: %s\n", secret.ID, secret.Name, secret.Type, payload.Metadata)
			for k, v := range payload.Fields {
				fmt.Printf("%s: %s\n", k, v)
			}
			if len(payload.Binary) > 0 {
				fmt.Printf("Binary: %d bytes\n", len(payload.Binary))
			}
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "secret id")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

func secretListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List secrets",
		Run: func(_ *cobra.Command, _ []string) {
			api, _, err := connectFromConfig()
			if err != nil {
				die(err)
			}
			defer api.Close()

			secrets, err := api.ListSecrets(context.Background())
			if err != nil {
				die(err)
			}
			fmt.Print(storage.FormatSecretList(secrets))
		},
	}
}

func secretDeleteCmd() *cobra.Command {
	var id string
	var version int64
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a secret",
		Run: func(_ *cobra.Command, _ []string) {
			api, _, err := connectFromConfig()
			if err != nil {
				die(err)
			}
			defer api.Close()

			if version == 0 {
				secret, err := api.GetSecret(context.Background(), id)
				if err != nil {
					die(err)
				}
				version = secret.Version
			}
			if err := api.DeleteSecret(context.Background(), id, version); err != nil {
				die(err)
			}
			fmt.Println("Deleted.")
		},
	}
	cmd.Flags().StringVar(&id, "id", "", "secret id")
	cmd.Flags().Int64Var(&version, "version", 0, "current version for optimistic locking")
	_ = cmd.MarkFlagRequired("id")
	return cmd
}

// BuildPayload constructs a secret payload from CLI flags.
func BuildPayload(secretType, login, pass, text, metadata, cardNumber string) (model.SecretType, *model.Payload, error) {
	payload := &model.Payload{Metadata: metadata, Fields: map[string]string{}}
	switch secretType {
	case "credentials":
		payload.Fields["login"] = login
		payload.Fields["password"] = pass
		return model.SecretTypeCredentials, payload, nil
	case "text":
		payload.Fields["text"] = text
		return model.SecretTypeText, payload, nil
	case "card":
		if err := validation.Luhn(cardNumber); err != nil {
			return 0, nil, err
		}
		payload.Fields["number"] = cardNumber
		return model.SecretTypeCard, payload, nil
	case "otp":
		payload.Fields["secret"] = text
		return model.SecretTypeOTP, payload, nil
	case "binary":
		payload.Binary = []byte(text)
		return model.SecretTypeBinary, payload, nil
	default:
		return 0, nil, fmt.Errorf("unknown secret type: %s", secretType)
	}
}
