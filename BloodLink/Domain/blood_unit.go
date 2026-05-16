package Domain

import "time"

type BloodUnit struct {
	BloodUnitID   string    `json:"blood_unit_id"`
	DonationID    string    `json:"donation_id,omitempty"`
	BloodType     string    `json:"blood_type"`
	ComponentType string    `json:"component_type"`
	QuantityML      int       `json:"quantity_ml"`
	CollectionDate time.Time `json:"collection_date"`
	ExpirationDate time.Time `json:"expiration_date"`
	Status        string    `json:"status"` // AVAILABLE | RESERVED | USED | EXPIRED
	IsDeleted     bool      `json:"is_deleted"`

	// Reservation fields (populated when status = RESERVED)
	ReservedForHospitalID string     `json:"reserved_for_hospital_id,omitempty"`
	ReservedAt            *time.Time `json:"reserved_at,omitempty"`
	RequestID             string     `json:"request_id,omitempty"`

	// Storage location fields
	StorageLocation string `json:"storage_location,omitempty"`
	RackNumber      string `json:"rack_number,omitempty"`
	ShelfNumber     string `json:"shelf_number,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

type BloodUnitFilter struct {
	BloodType     string `json:"blood_type"`
	ComponentType string `json:"component_type"`
	Status        string `json:"status"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	Quantity        int    `json:"quantity"`
	NearExpired   bool   `json:"near_expired"`
}

// ReservedUnitInfo is returned when units are reserved for a hospital request
type ReservedUnitInfo struct {
	BloodUnitID    string    `json:"blood_unit_id"`
	BloodType      string    `json:"blood_type"`
	QuantityML       int       `json:"quantity_ml"`
	ExpirationDate time.Time `json:"expiration_date"`
}

// ApproveRequestResult is the response returned when admin approves a blood request
type ApproveRequestResult struct {
	Status         string             `json:"status"`
	Message        string             `json:"message"`
	ReservedUnits  []ReservedUnitInfo `json:"reserved_units"`
	TotalQuantityML  int                `json:"total_quantity_ml"`
	RequestedCount int                `json:"requested_count"`
	FulfilledCount int                `json:"fulfilled_count"`
}

type ConvertCryoRequest struct {
	CryoprecipitateQuantity int  `json:"cryoprecipitate_quantity" binding:"required"`
	CryoPoorPlasmaQuantity  *int `json:"cryo_poor_plasma_quantity,omitempty"`
}

type InventoryListResponse struct {
	Total               int              `json:"total_blood_units"`
	Available           int              `json:"available_blood"`
	Reserved            int              `json:"reserved_blood"`
	Used                int              `json:"used_blood"`
	Expired             int              `json:"expired_blood"`
	NearExpired         int              `json:"near_expired_blood"`
	ByBloodType         map[string]int   `json:"by_blood_type"`
	ByComponentType     map[string]int   `json:"by_component_type"`
	ByBloodAndComponent map[string]int   `json:"by_blood_and_component"`
	Units               []BloodUnit      `json:"units"`
}