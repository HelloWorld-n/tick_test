package go_gin_pages

import (
	"fmt"
	"tick_test/types"
)

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
