package Domain

// BloodComponentInput is used in the lab technician request body
// to specify each component and its quantity when processing a donation.
type BloodComponentInput struct {
	ComponentType  string `json:"component_type"` // PRBC, PLATELETS, PLASMA, CRYOPRECIPITATE
	Quantity       int    `json:"quantity"`
	PositionNumber string `json:"position_number"` // newly added for slot assignment
}
