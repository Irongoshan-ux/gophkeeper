package cmd

import (
	"context"
	"fmt"

	"github.com/Irongoshan-ux/gophkeeper/internal/client"
	"github.com/Irongoshan-ux/gophkeeper/internal/client/storage"
	"github.com/Irongoshan-ux/gophkeeper/internal/config"
	"github.com/Irongoshan-ux/gophkeeper/pkg/crypto"
	"github.com/spf13/cobra"
)

func registerCmd() *cobra.Command {
	var login, password string
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user account",
		Run: func(_ *cobra.Command, _ []string) {
			api, err := client.Connect(client.Options{Address: serverAddr, Insecure: insecure})
			if err != nil {
				die(err)
			}
			defer api.Close()

			token, _, err := api.Register(context.Background(), login, password)
			if err != nil {
				die(err)
			}

			salt, err := crypto.NewSalt()
			if err != nil {
				die(err)
			}
			mp := masterPass
			if mp == "" {
				mp = password
			}

			cfg := &config.ClientConfig{
				ServerAddress: serverAddr,
				Token:         token,
				Salt:          salt,
			}
			if err := storage.SaveConfig(cfg); err != nil {
				die(err)
			}

			key, err := crypto.DeriveKey(mp, salt)
			if err != nil {
				die(err)
			}
			_ = key

			fmt.Println("Registered successfully.")
		},
	}
	cmd.Flags().StringVar(&login, "login", "", "account login")
	cmd.Flags().StringVar(&password, "password", "", "account password")
	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}

func loginCmd() *cobra.Command {
	var login, password string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the server",
		Run: func(_ *cobra.Command, _ []string) {
			api, err := client.Connect(client.Options{Address: serverAddr, Insecure: insecure})
			if err != nil {
				die(err)
			}
			defer api.Close()

			token, _, err := api.Login(context.Background(), login, password)
			if err != nil {
				die(err)
			}

			cfg, err := storage.LoadConfig()
			if err != nil {
				die(err)
			}
			cfg.ServerAddress = serverAddr
			cfg.Token = token
			if len(cfg.Salt) == 0 {
				salt, err := crypto.NewSalt()
				if err != nil {
					die(err)
				}
				cfg.Salt = salt
			}
			if err := storage.SaveConfig(cfg); err != nil {
				die(err)
			}

			if err := runSync(api, cfg); err != nil {
				die(err)
			}
			fmt.Println("Logged in and synced.")
		},
	}
	cmd.Flags().StringVar(&login, "login", "", "account login")
	cmd.Flags().StringVar(&password, "password", "", "account password")
	_ = cmd.MarkFlagRequired("login")
	_ = cmd.MarkFlagRequired("password")
	return cmd
}

func connectFromConfig() (*client.API, *config.ClientConfig, error) {
	cfg, err := storage.LoadConfig()
	if err != nil {
		return nil, nil, err
	}
	if cfg.Token == "" || cfg.ServerAddress == "" {
		return nil, nil, storage.ErrNotConfigured
	}
	addr := serverAddr
	if addr == "localhost:9090" && cfg.ServerAddress != "" {
		addr = cfg.ServerAddress
	}
	api, err := client.Connect(client.Options{
		Address:  addr,
		Insecure: insecure,
		Token:    cfg.Token,
	})
	if err != nil {
		return nil, nil, err
	}
	return api, cfg, nil
}

func deriveKeyFromConfig(cfg *config.ClientConfig) ([]byte, error) {
	mp := masterPass
	if mp == "" {
		return nil, crypto.ErrEmptyMasterPassword
	}
	return crypto.DeriveKey(mp, cfg.Salt)
}
