package substrate

import (
	"os/exec"

	_ "github.com/mattn/go-sqlite3"
)

type Process struct {
	Name string
	Cmd  *exec.Cmd
}

//func RunProcesses(processes []Process) error {
//	// ...
//}
