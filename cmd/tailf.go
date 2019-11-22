package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"sync"
	"tailer/lib"
)

func init() {
	rootCmd.AddCommand(tailfCmd)
}

var tailfCmd = &cobra.Command{
	Use:   "tailf projectName logFileAlias/logFile",
	Short: "similar to tailf",
	Long:  `similar to tailf`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		logFile := args[1]
		fmt.Println("project:", projectName)
		fmt.Println("log file:", logFile)
		servers := viper.GetStringSlice(fmt.Sprintf("project.%s.servers", projectName))
		fmt.Println("servers:", servers)
		if len(servers) == 0 {
			fmt.Println("no servers in project", projectName)
			return
		}
		paths := viper.GetStringMapString(fmt.Sprintf("project.%s.path", projectName))
		path, ok := paths[logFile]
		if ok {
			logFile = path
		}
		wg := &sync.WaitGroup{}
		for _, host := range servers {
			wg.Add(1)
			go func(host string) {
				defer wg.Done()
				session, err := lib.SshSession(host)
				if err != nil {
					fmt.Println(host, "err:", err)
					return
				}
				command := fmt.Sprintf("tailf %s", logFile)
				err = session.Start(command)
				if err != nil {
					fmt.Println(host, command, "got err:\n", err)
					return
				}
				err = session.Wait()
				if err != nil {
					fmt.Println(host, "err:", err)
					return
				}
			}(host)
		}
		wg.Wait()
	},
}
