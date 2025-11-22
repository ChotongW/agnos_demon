package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configFile string

var rootCmd = &cobra.Command{
	Use:   "demo-service",
	Short: "demo-service",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

// Execute root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default is config.yaml)")
}

func initConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath("cfg/")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.SetDefault("Log.Level", "debug")
	viper.SetDefault("Log.Color", true)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("unable to read config: %v\n", err)
		os.Exit(1)
	}
}
