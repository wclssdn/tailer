package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
	"tailer/lib"
	"time"

	"github.com/gookit/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(terminalCmd)
}

var terminalCmd = &cobra.Command{
	Use:   "terminal project_name",
	Short: "send command to all servers",
	Long:  `The output content is from all servers too.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		isSupportColor := color.IsSupportColor()
		projectName := args[0]
		servers := viper.GetStringSlice(fmt.Sprintf("project.%s.servers", projectName))
		if len(servers) == 0 {
			fmt.Println(color.Red.Render("no servers in project", projectName))
			return
		}
		fmt.Println("project:", color.Bold.Render(projectName))
		fmt.Println("servers:", servers)

		if !isSupportColor {
			fmt.Println("this terminal does not support color")
			time.Sleep(time.Millisecond * 10)
		}

		// 整体共用一个channel
		stdOutCh := make(chan []byte)
		stdErrCh := make(chan []byte)
		defer close(stdErrCh)
		defer close(stdOutCh)
		go func() {
			for {
				select {
				case buf := <-stdOutCh:
					if isSupportColor {
						fmt.Println(string(buf))
					} else {
						fmt.Println("--->", string(buf))
					}
				case buf := <-stdErrCh:
					if isSupportColor {
						fmt.Println(string(buf))
					} else {
						fmt.Println("--->", string(buf))
					}
				}
			}
		}()
		// 所有shell的统一输入
		inGroup := &sync.WaitGroup{}
		inGroup.Add(len(servers))
		stdIns := make([]io.WriteCloser, len(servers))

		wg := &sync.WaitGroup{}
		for i, host := range servers {
			wg.Add(1)
			go func(i int, host string) {
				defer wg.Done()
				session, err := lib.SshSession(host)
				if err != nil {
					fmt.Println(color.Cyan.Render(host), color.Red.Render(err))
					return
				}

				stdIn, err := session.StdinPipe()
				stdIns[i] = stdIn
				inGroup.Done()
				stdErr, err := session.StderrPipe()
				if err != nil {
					fmt.Println(color.Cyan.Render(host), color.Red.Render(err))
				}
				stdOut, err := session.StdoutPipe()
				if err != nil {
					fmt.Println(color.Cyan.Render(host), color.Red.Render(err))
				}

				err = session.Shell()
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
			}(i, host)
		}

		inGroup.Wait()
		fmt.Println("Terminal started. Write command and press Enter to execute on all servers.")

		go func() {
			scaner := bufio.NewScanner(os.Stdin)
			for {
				b := scaner.Scan()
				if !b {
					fmt.Println("scan failed, err:", scaner.Err())
					continue
				}
				by := scaner.Bytes()
				if len(by) == 0 {
					continue
				}
				fmt.Println("scan", len(by), "bytes", string(by), by)
				by = append(by, '\n')
				for _, stdIn := range stdIns {
					_, err := stdIn.Write(by)
					if err != nil {
						fmt.Println("write failed:", err)
						continue
					}
					//fmt.Println("write to", color.FgCyan.Render(servers[i]), n, "bytes")
				}
			}
		}()

		wg.Wait()
	},
}
