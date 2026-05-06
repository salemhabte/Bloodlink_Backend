package Jobs

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"bloodlink/Infrastructure"
	"bloodlink/Usecase"
	"fmt"
	"log"
	"time"
)

func StartExpirationJob(inventoryUC *Usecase.BloodInventoryUsecase, bloodReqRepo Interfaces.IBloodRequestRepository) {
	for {
		log.Println("[JOB] Running expiration and reservation cleanup...")

		// 1. Mark units past expiry date as EXPIRED
		if err := inventoryUC.UpdateExpiredUnits(); err != nil {
			log.Printf("[JOB ERROR] Failed to update expired units: %v", err)
		}

		// 2. Cleanup stale reservations (24h rule)
		cutoff := time.Now().Add(-24 * time.Hour)
		requestIDs, err := inventoryUC.ExpireReservations() // inventoryUC now has internal repo access
		if err != nil {
			log.Printf("[JOB ERROR] Failed to expire stale reservations: %v", err)
		}

		// 3. Reject associated blood requests
		for _, reqID := range requestIDs {
			log.Printf("[JOB] Auto-rejecting request %s due to stale reservation", reqID)
			
			// Mark request as REJECTED
			notes := "Automatically rejected: Blood reservation expired after 24 hours of waiting."
			_ = bloodReqRepo.UpdateRequestStatusWithDetails(reqID, Domain.BloodRequestStatusRejected, nil, notes, 0, 0)

			// Notify Hospital
			req, err := bloodReqRepo.GetRequestByID(reqID)
			if err == nil {
				go func() {
					subject := "Blood Request Auto-Rejected"
					content := fmt.Sprintf("Your request for %s blood has been automatically rejected because the reserved units were not collected within the required 24-hour window.", req.BloodType)
					_ = Infrastructure.SendBloodRequestNotification("hospitaladmin@bloodlink.com", subject, content)
				}()
			}
		}

		time.Sleep(1 * time.Hour) // Check more frequently than 24h to be accurate
	}
}