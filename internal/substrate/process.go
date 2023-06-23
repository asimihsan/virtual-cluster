package substrate

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/Netflix/go-expect"
	_ "github.com/mattn/go-sqlite3"
)

type ManagedProcess struct {
	Name        string
	Cmd         *exec.Cmd
	RunCommands []string
	Stop        chan struct{}
}

func runProcessAndStoreOutput(process *ManagedProcess, db *sql.DB) {
	sqliteLineWriter := NewLineWriter(func(line string) {
		fmt.Println("Writing line to DB:", line)
		_, err := db.Exec("INSERT INTO logs (process_name, output_type, content) VALUES (?, 'stdout', ?)", process.Name, line)
		if err != nil {
			log.Fatal(err)
		}
	})

	fmt.Println("Creating console")
	console, err := expect.NewConsole(
		expect.WithLogger(log.New(os.Stdout, "", 0)),
		expect.WithStdout(os.Stdout),
		expect.WithStdout(sqliteLineWriter),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer console.Close()

	cmd := exec.Command("bash")
	cmd.Stdin = console.Tty()
	cmd.Stdout = console.Tty()
	cmd.Stderr = console.Tty()

	var wg sync.WaitGroup
	wg.Add(1)

	fmt.Println("Starting process")
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Successfully started process, PID:", cmd.Process.Pid)

	for _, cmdStr := range process.RunCommands {
		fmt.Println("Running command:", cmdStr)
		time.Sleep(time.Second)
		console.SendLine(cmdStr)
		time.Sleep(time.Second)
	}

	console.SendLine("exit")

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait() // Wait for the goroutine to exit
}
