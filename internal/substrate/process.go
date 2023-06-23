package substrate

import (
	"bufio"
	"database/sql"
	"os/exec"

	_ "github.com/mattn/go-sqlite3"
)

func runProcessAndStoreOutput(processName string, cmd *exec.Cmd, db *sql.DB) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	// Start the process
	cmd.Start()

	// Read stdout/stderr and store them in the SQLite database
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			text := scanner.Text()
			// Store the output in the SQLite database
			db.Exec("INSERT INTO logs (process_name, output_type, content) VALUES (?, 'stdout', ?)", processName, text)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			text := scanner.Text()
			// Store the output in the SQLite database
			db.Exec("INSERT INTO logs (process_name, output_type, content) VALUES (?, 'stderr', ?)", processName, text)
		}
	}()

	// Wait for the process to finish and restart it if necessary
	cmd.Wait()
}
