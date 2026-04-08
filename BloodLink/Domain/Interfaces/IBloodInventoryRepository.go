package Domain

import "bloodlink/Domain"

type IBloodInventoryRepository interface {
	GetAllBloodUnits() ([]Domain.BloodUnit, error)
	GetBloodUnitByID(id string) (*Domain.BloodUnit, error)
	UpdateBloodUnitStatus(id string, status string) error
	DeleteBloodUnitByID(id string) error
	GetFullBloodUnitDetails(id string) (map[string]interface{}, error)
	FilterBloodUnits(unitID, bloodType, status, startDate, endDate string) ([]Domain.BloodUnit, error)
}