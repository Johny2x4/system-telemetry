package config

import (
	"fmt"
	"strconv"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/viper"
)

// AppConfig holds the settings saved in config.yaml
type AppConfig struct {
	Role            string `mapstructure:"role"`
	ServerURL       string `mapstructure:"server_url"`
	ListenPort      string `mapstructure:"listen_port"`
	DBPath          string `mapstructure:"db_path"`
	PollingInterval int    `mapstructure:"polling_interval"`
}

// LoadOrSetup reads the config file, or triggers the wizard if it's missing
func LoadOrSetup() (*AppConfig, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // Look in the current directory

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("No configuration found. Initializing Setup Wizard...")
			return RunSetupWizard()
		}
		return nil, err
	}

	var cfg AppConfig
	err := viper.Unmarshal(&cfg)
	return &cfg, err
}

// RunSetupWizard interactively asks the user how to configure the node.
// It is exported (capitalized) so we can force it via the CLI.
func RunSetupWizard() (*AppConfig, error) {
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

	// 3. Ask for Polling Interval
	promptInterval := &survey.Input{
		Message: "Enter the hardware polling interval in seconds (minimum 5):",
		Default: "10",
	}
	
	var intervalStr string
	survey.AskOne(promptInterval, &intervalStr, survey.WithValidator(func(val interface{}) error {
		str, ok := val.(string)
		if !ok {
			return fmt.Errorf("invalid input")
		}
		i, err := strconv.Atoi(str)
		if err != nil {
			return fmt.Errorf("please enter a valid number")
		}
		if i < 5 {
			return fmt.Errorf("polling interval must be at least 5 seconds to prevent system strain")
		}
		return nil
	}))
	
	cfg.PollingInterval, _ = strconv.Atoi(intervalStr)

	// 4. Save to config.yaml
	viper.Set("role", cfg.Role)
	viper.Set("server_url", cfg.ServerURL)
	viper.Set("listen_port", cfg.ListenPort)
	viper.Set("db_path", cfg.DBPath)
	viper.Set("polling_interval", cfg.PollingInterval)

	// THE FIX: Use WriteConfigAs to explicitly set the path and force an overwrite
	if err := viper.WriteConfigAs("config.yaml"); err != nil {
		return nil, fmt.Errorf("failed to save config.yaml: %w", err)
	}

	fmt.Println("\nConfiguration saved successfully to config.yaml!")
	return &cfg, nil
}