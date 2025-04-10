package repository

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

type IterationManipulator struct {
	Code        string                  `json:"Code"`
	Data        ManipulateIterationData `json:"Data"`
	Manipulator *time.Ticker            `json:"-"`
}

type ManipulateIterationData struct {
	Duration types.ISO8601Duration `json:"Duration" binding:"required"`
	Value    int                   `json:"Value" binding:"required"`
}

type UpdateIterationManipulatorData struct {
	Duration *types.ISO8601Duration `json:"Duration"`
	Value    *int                   `json:"Value"`
}

var IterationManipulatorMutex sync.Mutex
var IterationManipulators []*IterationManipulator = make([]*IterationManipulator, 0)

func loadIterationManipulators() error {
	if database != nil {
		if err := LoadIterationManipulatorsFromDatabase(); err != nil {
			return err
		}
	} else {
		if err := LoadIterationManipulatorsFromFile(); err != nil {
			return err
		}
	}
	for _, iterationManipulator := range IterationManipulators {
		dur, err := types.ParseISO8601Duration(iterationManipulator.Data.Duration, time.Second)
		if err != nil {
			return err
		}
		ticker := time.NewTicker(dur)
		iterationManipulator.Manipulator = ticker
		go ManipulateIteration(iterationManipulator)
	}
	return nil
}

func ApplyUpdateToIterationManipulator(data UpdateIterationManipulatorData, v *IterationManipulator) (dur time.Duration, err error) {
	if data.Duration != nil {
		dur, err = types.ParseISO8601Duration(*data.Duration, time.Second)
		if err != nil {
			return
		}
	}

	if data.Duration != nil {
		v.Data.Duration = *data.Duration
		v.Manipulator.Reset(dur)
	}
	if data.Value != nil {
		v.Data.Value = *data.Value
	}
	if database != nil {
		err = UpdateManipulatorInDatabase(v.Code, v.Data.Duration, v.Data.Value)
		if err != nil {
			return 0, err
		}
	}
	return
}

func ManipulateIteration(obj *IterationManipulator) error {
	for range obj.Manipulator.C {
		IterationMutex.Lock()
		Iteration += obj.Data.Value
		if err := SaveIteration(); err != nil {
			fmt.Printf("Error saving iteration: %v\n", err)
		}
		IterationMutex.Unlock()
	}
	return nil
}

func LoadIterationManipulatorsFromFile() error {
	manipulators, err := ReadManipulatorsFromFile()
	if err != nil {
		return err
	}
	IterationManipulators = manipulators
	return nil
}

func SaveIterationManipulators() error {
	IterationManipulatorMutex.Lock()
	defer IterationManipulatorMutex.Unlock()
	return WriteManipulatorsToFile(IterationManipulators)
}

func ReadManipulatorsFromFile() ([]*IterationManipulator, error) {
	file, err := os.Open(iterationManipulatorFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var manipulators []*IterationManipulator
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&manipulators); err != nil {
		return nil, err
	}
	return manipulators, nil
}

func WriteManipulatorsToFile(manipulators []*IterationManipulator) error {
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

func LoadIterationManipulatorsFromDatabase() error {
	query := `SELECT code, duration, value FROM manipulator`

	rows, err := database.Query(query)
	if err != nil {
		return err
	}
	defer rows.Close()
	IterationManipulators = make([]*IterationManipulator, 0)

	for rows.Next() {
		var code string
		var duration types.ISO8601Duration
		var value int

		if err := rows.Scan(&code, &duration, &value); err != nil {
			return err
		}

		iterationManipulator := &IterationManipulator{
			Code: code,
			Data: ManipulateIterationData{
				Duration: duration,
				Value:    value,
			},
		}
		IterationManipulators = append(IterationManipulators, iterationManipulator)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

func SaveIterationManipulatorToDatabase(obj *IterationManipulator) (err error) {
	if database == nil {
		return
	}
	query := `INSERT INTO manipulator (code, duration, value) VALUES ($1, $2, $3)`
	_, err = database.Exec(query, obj.Code, obj.Data.Duration, obj.Data.Value)
	return
}

func UpdateManipulatorInDatabase(code string, duration types.ISO8601Duration, value int) error {
	query := `UPDATE manipulator SET duration = $1, value = $2 WHERE code = $3`
	_, err := database.Exec(query, duration, value, code)
	return err
}

func DeleteManipulatorFromDatabase(code string) error {
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
