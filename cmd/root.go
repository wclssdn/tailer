package cmd

import (
	"fmt"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "tailer",
		Short: "A tool for interacting with multiple remote hosts at the same time.",
		Long: `A tool for interacting with multiple remote hosts at the same time. 
Especially it supports tail -f command.`,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tailer.yaml or /etc/.tailer)")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println("can't find home dir. use /etc instead.")
		}
		viper.AddConfigPath(home)
		viper.AddConfigPath("/etc")
		viper.SetConfigName(".tailer")
	}

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("read config file failed. err:", err)
	}
}
