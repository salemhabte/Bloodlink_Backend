package Domain

import (
	"bloodlink/Domain"
	"time"
)

type IBloodInventoryRepository interface {
	GetAllBloodUnits(filter Domain.BloodUnitFilter) ([]Domain.BloodUnit, error)
	GetBloodUnitByID(id string) (*Domain.BloodUnit, error)
	UpdateBloodUnitStatus(id string, status string) error
	DeleteBloodUnitByID(id string) error
	GetFullBloodUnitDetails(id string) (map[string]interface{}, error)
	FilterBloodUnits(filter Domain.BloodUnitFilter) ([]Domain.BloodUnit, error)
	MarkExpiredUnits() error
	CountAvailableUnitsByBloodType(bloodType string) (int, error)
	ConsumeUnits(bloodType string, quantity int) error

	// Reservation workflow
	ReserveUnitsForHospital(bloodType string, componentType string, quantity int, hospitalID string, requestID string) ([]Domain.BloodUnit, error)
	MarkUnitAsUsed(unitID string) error
	ExpireStaleReservations(cutoff time.Time) ([]string, error) // returns affected request_ids
	GetReservedUnitsByHospitalID(hospitalID string) ([]Domain.BloodUnit, error)
	GetReservedUnitsByRequestID(requestID string) ([]Domain.BloodUnit, error)

	// Delete with audit
	DeleteWithAudit(unitID string) error

	ConvertPlasmaToCryo(plasmaUnitID string, cryo *Domain.BloodUnit, cryoPoor *Domain.BloodUnit) error
}