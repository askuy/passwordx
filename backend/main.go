package main

import (
	"github.com/askuy/passwordx/backend/cmd"
	_ "github.com/askuy/passwordx/backend/cmd/init"
	_ "github.com/askuy/passwordx/backend/cmd/server"
	"github.com/gotomicro/ego/core/elog"
)

func main() {
	err := cmd.RootCommand.Execute()
	if err != nil {
		elog.Panic("startup", elog.FieldErr(err))
	}
}
