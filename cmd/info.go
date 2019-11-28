package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"os"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "some useful information",
	Long:  `some useful information`,
	Run: func(cmd *cobra.Command, args []string) {
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println("get home dir failed. err:", err)
		}
		fmt.Println("home is", home)
		fmt.Println("loaded config file:", viper.ConfigFileUsed())
		fmt.Println("TERM:", os.Getenv("TERM"))
		fmt.Println("ConEmuANSI:", os.Getenv("ConEmuANSI"))
		fmt.Println("ANSICON:", os.Getenv("ANSICON"))
	},
}
