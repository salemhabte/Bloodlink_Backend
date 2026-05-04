package Domain

import "time"

type DonorBloodRequest struct {
	RequestID     string
	DonorID       string
	BloodType     string
	QuantityML    int
	Reason        string
	PriorityScore int
	Status        string
	CreatedAt     time.Time
}