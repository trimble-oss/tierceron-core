package flow

import (
	"log"
)

type ShellContext interface {
	GetEnv() string
	GetLogger() *log.Logger
	Log(string, error)
}

type ShellMachineContext interface {
	NewShellContext(string, log.Logger) ShellContext
	ExecTrcShellCommand(ShellContext, string, []string) (any, error)
}
