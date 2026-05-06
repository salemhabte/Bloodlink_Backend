package Domain

import (
	"bloodlink/Domain"
	"time"
)

type IBloodInventoryRepository interface {
	GetAllBloodUnits() ([]Domain.BloodUnit, error)
	GetBloodUnitByID(id string) (*Domain.BloodUnit, error)
	UpdateBloodUnitStatus(id string, status string) error
	DeleteBloodUnitByID(id string) error
	GetFullBloodUnitDetails(id string) (map[string]interface{}, error)
	FilterBloodUnits(unitID, bloodType, status, startDate, endDate string) ([]Domain.BloodUnit, error)
	MarkExpiredUnits() error
	CountAvailableUnitsByBloodType(bloodType string) (int, error)
	ConsumeUnits(bloodType string, quantity int) error

	// Reservation workflow
	ReserveUnitsForHospital(bloodType string, quantity int, hospitalID string, requestID string) ([]Domain.BloodUnit, error)
	MarkUnitAsUsed(unitID string) error
	ExpireStaleReservations(cutoff time.Time) ([]string, error) // returns affected request_ids
	GetReservedUnitsByHospitalID(hospitalID string) ([]Domain.BloodUnit, error)

	// Delete with audit
	DeleteWithAudit(unitID string) error
}