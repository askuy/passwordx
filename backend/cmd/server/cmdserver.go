package server

import (
	"github.com/askuy/passwordx/backend/cmd"
	"github.com/spf13/cobra"
)

var CmdRun = &cobra.Command{
	Use:   "server",
	Short: "start passwordx server",
	Long:  `start passwordx server`,
	Run:   ServerFunc,
}

func init() {
	cmd.RootCommand.AddCommand(CmdRun)
}

func ServerFunc(cmd *cobra.Command, args []string) {
	// 初始化 EGO 应用
	Server()
}
