package Repository

import (
	"context"
	"database/sql"
	"errors"
	"log"

	domain "bloodlink/Domain"
	domainInterface "bloodlink/Domain/Interfaces"
)

type HospitalRepository struct {
	DB *sql.DB
}

func NewHospitalRepository(db *sql.DB) domainInterface.IHospitalRepository {
	return &HospitalRepository{DB: db}
}

func (r *HospitalRepository) Create(ctx context.Context, hospital *domain.Hospital) error {
	query := `INSERT INTO hospitals (hospital_id, hospital_name, address, city, phone, contact_person_name, contact_person_phone, status, created_at) 
               VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.DB.ExecContext(ctx, query,
		hospital.HospitalID,
		hospital.HospitalName,
		hospital.Address,
		hospital.City,
		hospital.Phone,
		hospital.ContactPersonName,
		hospital.ContactPersonPhone,
		hospital.Status,
		hospital.CreatedAt,
	)

	if err != nil {
		log.Printf("[DATABASE ERROR] CreateHospital failed: %v", err)
		return err
	}

	return nil
}

func (r *HospitalRepository) GetByID(ctx context.Context, id string) (*domain.Hospital, error) {
	query := `SELECT hospital_id, hospital_name, address, COALESCE(city, ''), phone, COALESCE(contact_person_name, ''), COALESCE(contact_person_phone, ''), status, COALESCE(document1_url, ''), COALESCE(document2_url, ''), created_at 
              FROM hospitals WHERE hospital_id = ?`

	row := r.DB.QueryRowContext(ctx, query, id)

	var h domain.Hospital
	err := row.Scan(
		&h.HospitalID,
		&h.HospitalName,
		&h.Address,
		&h.City,
		&h.Phone,
		&h.ContactPersonName,
		&h.ContactPersonPhone,
		&h.Status,
		&h.Document1URL,
		&h.Document2URL,
		&h.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("hospital not found")
		}
		return nil, err
	}

	return &h, nil
}

func (r *HospitalRepository) Update(ctx context.Context, h *domain.Hospital) error {
	query := `UPDATE hospitals 
              SET hospital_name = ?, address = ?, city = ?, phone = ?, contact_person_name = ?, contact_person_phone = ?, status = ? 
              WHERE hospital_id = ?`

	_, err := r.DB.ExecContext(ctx, query,
		h.HospitalName,
		h.Address,
		h.City,
		h.Phone,
		h.ContactPersonName,
		h.ContactPersonPhone,
		h.Status,
		h.HospitalID,
	)

	if err != nil {
		log.Printf("[DATABASE ERROR] UpdateHospital failed: %v", err)
		return err
	}

	return nil
}

func (r *HospitalRepository) UpdateDocuments(ctx context.Context, id string, doc1 string, doc2 string) error {
	query := `UPDATE hospitals SET document1_url = ?, document2_url = ? WHERE hospital_id = ?`
	_, err := r.DB.ExecContext(ctx, query, doc1, doc2, id)
	if err != nil {
		log.Printf("[DATABASE ERROR] UpdateDocuments failed: %v", err)
		return err
	}
	return nil
}

