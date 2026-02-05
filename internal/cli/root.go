package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "carnie",
		Aliases: []string{"cn"},
		Short:   "Carnie agent orchestration",
		Long:    "Carnie is a reliable, portable agent orchestration system.",
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.carnie.yaml)")
	cobra.OnInitialize(initConfig)

	rootCmd.AddCommand(newDashboardCommand())
	rootCmd.AddCommand(newCampCommand())
	rootCmd.AddCommand(newOperatorCommand())
	rootCmd.AddCommand(newPrimeCommand())
	rootCmd.AddCommand(newWorkOrderCommand())

	return rootCmd
}

func Execute() error {
	return NewRootCommand().Execute()
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "unable to resolve home directory")
			return
		}

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".carnie")
	}

	viper.SetEnvPrefix("cn")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		viper.Set("config_file", filepath.Base(viper.ConfigFileUsed()))
	}
}
