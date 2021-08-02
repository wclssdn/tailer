package cmd

import (
	"fmt"
	"os"

	"github.com/gookit/color"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

		if color.IsSupportColor() {
			fmt.Println("color: " + color.FgLightGreen.Render("support"))
		} else {
			fmt.Println("color: not support")
		}
		if color.IsSupport256Color() {
			fmt.Println("256 color: " + color.FgLightGreen.Render("support"))
		} else {
			if color.IsSupportColor() {
				fmt.Println("256 color: " + color.FgLightRed.Render("not support"))
			} else {
				fmt.Println("256 color: not support")
			}
		}
	},
}
