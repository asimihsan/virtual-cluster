package substrate

import (
	"database/sql"
	"fmt"
	"log"
	"os/exec"
	"sync"

	"github.com/asimihsan/virtual-cluster/internal/parser"
	_ "github.com/mattn/go-sqlite3"
)

type Manager struct {
	dbPath    string
	db        *sql.DB
	processes []*ManagedProcess
	wg        sync.WaitGroup
}

func NewManager(dbPath string) (*Manager, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec("PRAGMA synchronous = FULL")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY,
			process_name TEXT,
			output_type TEXT,
			content TEXT
		)
	`)
	if err != nil {
		return nil, err
	}

	return &Manager{
		dbPath: dbPath,
		db:     db,
	}, nil
}

func (m *Manager) Close() error {
	m.StopAllProcesses()
	return m.db.Close()
}

// takes the combined slice of VClusterServiceDefinitionAST and VClusterDependencyDefinitionAST and
// performs a topological sort.
func topologicalSort(definitions []interface{}) ([]interface{}, error) {
	// Perform topological sort on the combined slice

	return definitions, nil
}

func (m *Manager) StartServicesAndDependencies(asts []*parser.VClusterAST) error {
	// Combine and check for duplicate names
	// Perform topological sort
	// Start services and dependencies one at a time

	for _, ast := range asts {
		for _, service := range ast.Services {
			fmt.Println("Starting service:", service.Name)
			process := &ManagedProcess{
				Name:        service.Name,
				Cmd:         exec.Command("bash"),
				RunCommands: service.RunCommands,
				Stop:        make(chan struct{}),
			}
			m.processes = append(m.processes, process)
			m.wg.Add(1)
			go runProcessAndStoreOutput(process, m.db)
			fmt.Println("Started service:", service.Name)
		}
	}

	m.wg.Wait() // Wait for all processes to finish

	return nil
}

func (m *Manager) StopAllProcesses() {
	for _, process := range m.processes {
		process.Stop <- struct{}{}
	}
}
