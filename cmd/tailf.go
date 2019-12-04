package cmd

import (
	"bufio"
	"fmt"
	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
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
		servers := viper.GetStringSlice(fmt.Sprintf("project.%s.servers", projectName))
		if len(servers) == 0 {
			fmt.Println(color.Red.Render("no servers in project: ", projectName))
			return
		}
		paths := viper.GetStringMapString(fmt.Sprintf("project.%s.path", projectName))
		path, ok := paths[logFile]
		if ok {
			logFile = path
		}
		fmt.Println("project:", color.Bold.Render(projectName))
		fmt.Println("log file:", color.Bold.Render(logFile))
		fmt.Println("servers:", servers)

		// 整体共用一个channel
		stdOutCh := make(chan []byte)
		stdErrCh := make(chan []byte)
		defer close(stdErrCh)
		defer close(stdOutCh)
		go func() {
			for {
				select {
				case buf := <-stdOutCh:
					if color.IsSupportColor() {
						fmt.Println(string(buf))
					} else {
						fmt.Println("--->", string(buf))
					}
				case buf := <-stdErrCh:
					if color.IsSupportColor() {
						fmt.Println(string(buf))
					} else {
						fmt.Println("--->", string(buf))
					}
				}
			}
		}()

		wg := &sync.WaitGroup{}
		for _, host := range servers {
			wg.Add(1)
			go func(host string) {
				defer wg.Done()
				session, err := lib.SshSession(host)
				if err != nil {
					fmt.Println(color.Cyan.Render(host), color.Red.Render(err))
					return
				}
				command := fmt.Sprintf("tailf %s", logFile)

				stdErr, err := session.StderrPipe()
				if err != nil {
					fmt.Println(color.Cyan.Render(host), color.Red.Render(err))
				}
				stdOut, err := session.StdoutPipe()
				if err != nil {
					fmt.Println(color.Cyan.Render(host), color.Red.Render(err))
				}

				err = session.Start(command)
				if err != nil {
					fmt.Println(color.Cyan.Render(host), color.Red.Render(err))
					return
				}

				go func() {
					re := bufio.NewReader(stdErr)
					bigBuf := make([]byte, 0)
					for {
						buf, isP, err := re.ReadLine()
						if isP {
							bigBuf = append(bigBuf, buf...)
							fmt.Println(color.Cyan.Render(host), color.Red.Render("todo with isPrefix is true"), "len:", len(buf), "cap:", cap(buf))
							continue
						}
						if len(bigBuf) > 0 {
							bigBuf = append(bigBuf, buf...)
						} else {
							bigBuf = buf
						}
						if err != nil {
							stdErrCh <- []byte(color.FgCyan.Render(host) + " " + color.FgRed.Render(err))
							if err == io.EOF {
								// todo retry after file created
								break
							}
						} else {
							stdErrCh <- []byte(color.FgCyan.Render(host) + " " + color.FgRed.Render(string(bigBuf)))
						}
						// reset
						bigBuf = make([]byte, 0)
					}
				}()

				go func() {
					re := bufio.NewReader(stdOut)
					bigBuf := make([]byte, 0)
					for {
						buf, isP, err := re.ReadLine()
						if isP {
							bigBuf = append(bigBuf, buf...)
							fmt.Println(color.Cyan.Render(host), color.Red.Render("todo with isPrefix is true"), "len:", len(buf), "cap:", cap(buf))
							continue
						}
						if len(bigBuf) > 0 {
							bigBuf = append(bigBuf, buf...)
						} else {
							bigBuf = buf
						}
						if err != nil {
							stdOutCh <- []byte(color.FgCyan.Render(host) + " " + color.FgRed.Render(err))
							if err == io.EOF {
								// todo retry after file created
								break
							}
						} else {
							stdOutCh <- []byte(color.FgCyan.Render(host) + " " + string(bigBuf))
						}
						// reset
						bigBuf = make([]byte, 0)
					}
				}()

				err = session.Wait()
				if err != nil {
					fmt.Println(color.Cyan.Render(host), color.Red.Render(err))
					return
				}
			}(host)
		}
		wg.Wait()
	},
}
