/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/peterh/liner"
	"github.com/spf13/cobra"
)

type shell struct {
	historyFile string
}

// shellCmd represents the shell command
var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "open shell mode",
	Long:  `open shell mode`,
	Run:   runShell,
}

func init() {
	rootCmd.AddCommand(shellCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// shellCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// shellCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runShell(cmd *cobra.Command, args []string) {
	fmt.Println("shell called")
	s := &shell{historyFile: historyFilePath()}
	s.run()
}

func (s *shell) run() {
	lnr := liner.NewLiner()
	lnr.SetCtrlCAborts(true)

	// load history before start of the prmopt
	s.loadHistory(lnr)
	defer func() {
		s.saveHistory(lnr)
		_ = lnr.Close()
		fmt.Println("bye!")
	}()

	for {
		pmt, err := lnr.Prompt("poc>")
		if err != nil {
			printError(err)
			if err == io.EOF || err == liner.ErrPromptAborted {
				break
			}
		}

		pmt = strings.TrimSpace(pmt)
		if pmt == "" {
			continue
		}

		lnr.AppendHistory(pmt)
		_, cancel := context.WithCancel(context.Background())
		go func() {
			sigChan := make(chan os.Signal, 1)
			defer signal.Stop(sigChan)

			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
			select {
			case <-sigChan:
				// cancel context or any other callback function.
				cancel()
			}
		}()

		// perform operation
		// print response to std out.
		fmt.Printf("response : %v\n", pmt)
	}
}

func (s *shell) loadHistory(lnr *liner.State) {
	f, err := os.OpenFile(s.historyFile, os.O_RDONLY|os.O_CREATE, 0640)
	if err != nil {
		printError(err)
		return
	}

	defer f.Close()
	_, err = lnr.ReadHistory(f)
	if err != nil {
		printError(err)
	}
}

func (s *shell) saveHistory(lnr *liner.State) {
	f, err := os.OpenFile(s.historyFile, os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		printError(err)
		return
	}
	defer f.Close()

	_, err = lnr.WriteHistory(f)
	if err != nil {
		printError(err)
	}

}

func historyFilePath() string {
	var fileDir = os.TempDir()
	usr, err := user.Current()
	if err == nil {
		fileDir = usr.HomeDir
	}
	return filepath.Join(fileDir, ".poc_history")
}

func printError(err error) {
	fmt.Fprintln(os.Stderr, err)
}
