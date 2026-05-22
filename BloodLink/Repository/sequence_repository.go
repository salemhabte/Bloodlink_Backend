package Repository

import (
	"database/sql"
)

type sequenceRepository struct {
	db *sql.DB
}

func NewSequenceRepository(db *sql.DB) *sequenceRepository {
	return &sequenceRepository{db: db}
}

func (r *sequenceRepository) GetNextSequenceValue(counterType string, year int) (int, error) {
	query := `
		INSERT INTO id_counters (counter_type, counter_year, current_value)
		VALUES ($1, $2, 1)
		ON CONFLICT (counter_type, counter_year)
		DO UPDATE SET current_value = id_counters.current_value + 1
		RETURNING current_value
	`
	var nextVal int
	err := r.db.QueryRow(query, counterType, year).Scan(&nextVal)
	if err != nil {
		return 0, err
	}
	return nextVal, nil
}
