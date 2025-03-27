package go_gin_pages

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"tick_test/types"
	"time"
)

const iterationManipulatorFile = "../.data/IterationManipulators.json"

type iterationManipulator struct {
	Code        string                  `json:"Code"`
	Data        manipulateIterationData `json:"Data"`
	Manipulator *time.Ticker            `json:"-"`
}

var iterationManipulatorMutex sync.Mutex

func loadIterationManipulators() error {
	if database != nil {
		if err := loadIterationManipulatorsFromDatabase(); err != nil {
			return err
		}
	} else {
		if err := loadIterationManipulatorsFromFile(); err != nil {
			return err
		}
	}
	for _, iterationManipulator := range iterationManipulators {
		dur, err := parseISO8601Duration(iterationManipulator.Data.Duration, time.Second)
		if err != nil {
			return err
		}
		ticker := time.NewTicker(dur)
		iterationManipulator.Manipulator = ticker
		go manipulateIteration(iterationManipulator)
	}
	return nil
}

func loadIterationManipulatorsFromFile() error {
	manipulators, err := readManipulatorsFromFile()
	if err != nil {
		return err
	}
	iterationManipulators = manipulators
	return nil
}

func saveIterationManipulators() error {
	iterationManipulatorMutex.Lock()
	defer iterationManipulatorMutex.Unlock()
	return writeManipulatorsToFile(iterationManipulators)
}

var iterationManipulators []*iterationManipulator = make([]*iterationManipulator, 0)

func readManipulatorsFromFile() ([]*iterationManipulator, error) {
	file, err := os.Open(iterationManipulatorFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var manipulators []*iterationManipulator
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&manipulators); err != nil {
		return nil, err
	}
	return manipulators, nil
}

func writeManipulatorsToFile(manipulators []*iterationManipulator) error {
	if err := os.MkdirAll(filepath.Dir(iterationManipulatorFile), 0755); err != nil {
		return err
	}

	file, err := os.Create(iterationManipulatorFile)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	return encoder.Encode(manipulators)
}

func loadIterationManipulatorsFromDatabase() error {
	query := `SELECT code, duration, value FROM manipulator`

	rows, err := database.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	iterationManipulators = make([]*iterationManipulator, 0)

	for rows.Next() {
		var code string
		var duration types.ISO8601Duration
		var value int

		if err := rows.Scan(&code, &duration, &value); err != nil {
			return err
		}

		iterationManipulator := &iterationManipulator{
			Code: code,
			Data: manipulateIterationData{
				Duration: duration,
				Value:    value,
			},
		}
		iterationManipulators = append(iterationManipulators, iterationManipulator)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func saveIterationManipulatorToDatabase(obj *iterationManipulator) (err error) {
	if database == nil {
		return
	}
	query := `INSERT INTO manipulator (code, duration, value) VALUES ($1, $2, $3)`
	_, err = database.Exec(query, obj.Code, obj.Data.Duration, obj.Data.Value)
	return
}

func updateManipulatorInDatabase(code string, duration types.ISO8601Duration, value int) error {
	query := `UPDATE manipulator SET duration = $1, value = $2 WHERE code = $3`
	_, err := database.Exec(query, duration, value, code)
	return err
}

func deleteManipulatorFromDatabase(code string) error {
	query := `DELETE FROM manipulator WHERE code = $1`
	_, err := database.Exec(query, code)
	return err
}

func doPostgresPreparationForManipulator() {
	if database != nil {
		result, _ := database.Query(`
			CREATE TABLE IF NOT EXISTS manipulator (
				code varchar(100) PRIMARY KEY,
				duration varchar(30) NOT NULL,
				value integer NOT NULL
			);
		`)
		fmt.Println(result)
	}
}
