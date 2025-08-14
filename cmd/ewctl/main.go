package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/config"
	"github.com/thalysguimaraes/elementor-whatsapp/internal/tui"
)

var (
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	cfgFile   string
	debugMode bool
)

var rootCmd = &cobra.Command{
	Use:   "ewctl",
	Short: "Elementor WhatsApp Manager - Manage webhooks and contacts",
	Long: `ewctl is a terminal user interface for managing Elementor WhatsApp webhooks.
	
It provides an intuitive interface to:
- Manage webhook forms and their configurations
- Organize contacts and recipients
- Test webhook endpoints
- Export and import configurations`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if debugMode {
			log.SetLevel(log.DebugLevel)
		}

		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return tui.Run(cfg)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ewctl %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/ewctl/config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug mode")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(formsCmd())
	rootCmd.AddCommand(contactsCmd())
	rootCmd.AddCommand(webhookCmd())
	rootCmd.AddCommand(configCmd())
}

func initConfig() {
	if debugMode {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal("Error", "err", err)
		os.Exit(1)
	}
}

func formsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forms",
		Short: "Manage webhook forms",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all forms",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			return tui.RunFormsListView(cfg)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create a new form",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			return tui.RunFormCreateView(cfg)
		},
	})

	return cmd
}

func contactsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contacts",
		Short: "Manage contacts",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all contacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			return tui.RunContactsListView(cfg)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "add",
		Short: "Add a new contact",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			return tui.RunContactCreateView(cfg)
		},
	})

	return cmd
}

func webhookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "webhook",
		Short: "Test webhook endpoints",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "test [form-id]",
		Short: "Test a webhook endpoint",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			
			var formID string
			if len(args) > 0 {
				formID = args[0]
			}
			
			return tui.RunWebhookTestView(cfg, formID)
		},
	})

	return cmd
}

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			return config.Print(cfg)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "edit",
		Short: "Edit configuration interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return err
			}
			return tui.RunConfigEditView(cfg)
		},
	})

	return cmd
}