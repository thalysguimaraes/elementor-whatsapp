package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Cloudflare CloudflareConfig `yaml:"cloudflare" mapstructure:"cloudflare"`
	ZAPI       ZAPIConfig       `yaml:"zapi" mapstructure:"zapi"`
	UI         UIConfig         `yaml:"ui" mapstructure:"ui"`
	Profiles   map[string]Profile `yaml:"profiles,omitempty" mapstructure:"profiles"`
}

type CloudflareConfig struct {
	AccountID  string `yaml:"account_id" mapstructure:"account_id"`
	APIToken   string `yaml:"api_token" mapstructure:"api_token"`
	DatabaseID string `yaml:"database_id" mapstructure:"database_id"`
	WorkerURL  string `yaml:"worker_url" mapstructure:"worker_url"`
}

type ZAPIConfig struct {
	InstanceID    string `yaml:"instance_id" mapstructure:"instance_id"`
	InstanceToken string `yaml:"instance_token" mapstructure:"instance_token"`
	ClientToken   string `yaml:"client_token" mapstructure:"client_token"`
}

type UIConfig struct {
	Theme              string        `yaml:"theme" mapstructure:"theme"`
	Mouse              bool          `yaml:"mouse" mapstructure:"mouse"`
	Animations         bool          `yaml:"animations" mapstructure:"animations"`
	VimBindings        bool          `yaml:"vim_bindings" mapstructure:"vim_bindings"`
	AutoRefresh        time.Duration `yaml:"auto_refresh" mapstructure:"auto_refresh"`
	ConfirmDestructive bool          `yaml:"confirm_destructive" mapstructure:"confirm_destructive"`
}

type Profile struct {
	WorkerURL string `yaml:"worker_url,omitempty" mapstructure:"worker_url"`
}

func DefaultConfig() *Config {
	return &Config{
		Cloudflare: CloudflareConfig{
			WorkerURL: "https://elementor-whatsapp.workers.dev",
		},
		UI: UIConfig{
			Theme:              "charm",
			Mouse:              true,
			Animations:         true,
			VimBindings:        false,
			AutoRefresh:        30 * time.Second,
			ConfirmDestructive: true,
		},
		Profiles: map[string]Profile{
			"dev": {
				WorkerURL: "http://localhost:8787",
			},
			"staging": {
				WorkerURL: "https://staging.workers.dev",
			},
		},
	}
}

func Load(configFile string) (*Config, error) {
	cfg := DefaultConfig()

	v := viper.New()
	v.SetConfigType("yaml")
	
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		configDir, err := getConfigDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get config dir: %w", err)
		}
		
		v.AddConfigPath(configDir)
		v.SetConfigName("config")
		
		// Create config dir if it doesn't exist
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create config dir: %w", err)
		}
	}

	// Set environment variable support
	v.SetEnvPrefix("EWCTL")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Debug("Config file not found, using defaults")
			// Check for legacy .env file in manager directory
			if err := loadLegacyEnv(cfg); err != nil {
				log.Debug("No legacy .env file found")
			}
		} else {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
	} else {
		if err := v.Unmarshal(cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	// Override with environment variables
	if accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID"); accountID != "" {
		cfg.Cloudflare.AccountID = accountID
	}
	if apiToken := os.Getenv("CLOUDFLARE_API_TOKEN"); apiToken != "" {
		cfg.Cloudflare.APIToken = apiToken
	}
	if databaseID := os.Getenv("DATABASE_ID"); databaseID != "" {
		cfg.Cloudflare.DatabaseID = databaseID
	}
	if workerURL := os.Getenv("WORKER_URL"); workerURL != "" {
		cfg.Cloudflare.WorkerURL = workerURL
	}
	if instanceID := os.Getenv("ZAPI_INSTANCE_ID"); instanceID != "" {
		cfg.ZAPI.InstanceID = instanceID
	}
	if instanceToken := os.Getenv("ZAPI_INSTANCE_TOKEN"); instanceToken != "" {
		cfg.ZAPI.InstanceToken = instanceToken
	}
	if clientToken := os.Getenv("ZAPI_CLIENT_TOKEN"); clientToken != "" {
		cfg.ZAPI.ClientToken = clientToken
	}

	return cfg, nil
}

func Save(cfg *Config, configFile string) error {
	var configPath string
	if configFile != "" {
		configPath = configFile
	} else {
		configDir, err := getConfigDir()
		if err != nil {
			return fmt.Errorf("failed to get config dir: %w", err)
		}
		configPath = filepath.Join(configDir, "config.yaml")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	log.Info("Configuration saved", "path", configPath)
	return nil
}

func Print(cfg *Config) error {
	// Mask sensitive data
	masked := *cfg
	if masked.Cloudflare.APIToken != "" {
		masked.Cloudflare.APIToken = maskString(masked.Cloudflare.APIToken)
	}
	if masked.ZAPI.InstanceToken != "" {
		masked.ZAPI.InstanceToken = maskString(masked.ZAPI.InstanceToken)
	}
	if masked.ZAPI.ClientToken != "" {
		masked.ZAPI.ClientToken = maskString(masked.ZAPI.ClientToken)
	}

	data, err := yaml.Marshal(masked)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".config", "ewctl"), nil
}

func loadLegacyEnv(cfg *Config) error {
	// Try to load from manager/.env for backward compatibility
	envPath := filepath.Join("manager", ".env")
	if _, err := os.Stat(envPath); err != nil {
		return err
	}

	viper.SetConfigFile(envPath)
	viper.SetConfigType("env")
	
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	cfg.Cloudflare.AccountID = viper.GetString("CLOUDFLARE_ACCOUNT_ID")
	cfg.Cloudflare.APIToken = viper.GetString("CLOUDFLARE_API_TOKEN")
	cfg.Cloudflare.DatabaseID = viper.GetString("DATABASE_ID")
	cfg.Cloudflare.WorkerURL = viper.GetString("WORKER_URL")

	log.Info("Migrated configuration from legacy .env file")
	return nil
}

func maskString(s string) string {
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}

func (c *Config) Validate() error {
	if c.Cloudflare.AccountID == "" {
		return fmt.Errorf("cloudflare.account_id is required")
	}
	if c.Cloudflare.APIToken == "" {
		return fmt.Errorf("cloudflare.api_token is required")
	}
	if c.Cloudflare.DatabaseID == "" {
		return fmt.Errorf("cloudflare.database_id is required")
	}
	if c.Cloudflare.WorkerURL == "" {
		return fmt.Errorf("cloudflare.worker_url is required")
	}
	return nil
}