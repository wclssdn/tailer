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
		Short: "A tool for tail log file",
		Long: `Tailer is a tool for tail log file.
Servers are retrieved from HULK.`,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.tailer.yaml)")
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
			home = "/etc"
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".tailer")
	}

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("read config file failed. err:", err)
	}
}
