package config

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/viper"
)

// AppConfig holds the settings saved in config.yaml
type AppConfig struct {
	Role       string `mapstructure:"role"`        // "Client" or "Server"
	ServerURL  string `mapstructure:"server_url"`  // Used by Client to know where to send data
	ListenPort string `mapstructure:"listen_port"` // Used by Server to open the API port
	DBPath     string `mapstructure:"db_path"`     // Used by Server for SQLite
}

// LoadOrSetup reads the config file, or triggers the wizard if it's missing
func LoadOrSetup() (*AppConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // Look in the current directory

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No configuration found. Initializing Setup Wizard...")
			return runSetupWizard()
		}
		return nil, err
	}

	var cfg AppConfig
	err := viper.Unmarshal(&cfg)
	return &cfg, err
}

// runSetupWizard interactively asks the user how to configure the node
func runSetupWizard() (*AppConfig, error) {
	var cfg AppConfig

	// 1. Ask for the Role
	promptRole := &survey.Select{
		Message: "Initialize this node as a Client or Server?",
		Options: []string{"Client", "Server"},
	}
	survey.AskOne(promptRole, &cfg.Role)

	// 2. Ask role-specific questions
	if cfg.Role == "Client" {
		promptURL := &survey.Input{
			Message: "Enter the Server IP and Port (e.g., http://192.168.1.100:8080):",
		}
		survey.AskOne(promptURL, &cfg.ServerURL)
	} else {
		promptPort := &survey.Input{
			Message: "Enter the port for the Server API to listen on:",
			Default: "8080",
		}
		survey.AskOne(promptPort, &cfg.ListenPort)
		
		promptDB := &survey.Input{
			Message: "Enter the filename for the SQLite database:",
			Default: "telemetry.db",
		}
		survey.AskOne(promptDB, &cfg.DBPath)
	}

	// 3. Save to config.yaml
	viper.Set("role", cfg.Role)
	viper.Set("server_url", cfg.ServerURL)
	viper.Set("listen_port", cfg.ListenPort)
	viper.Set("db_path", cfg.DBPath)

	if err := viper.SafeWriteConfig(); err != nil {
		return nil, fmt.Errorf("failed to save config.yaml: %w", err)
	}

	fmt.Println("\nConfiguration saved successfully to config.yaml!")
	return &cfg, nil
}