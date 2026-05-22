package jobs

import (
	"bloodlink/Domain"
	Interfaces "bloodlink/Domain/Interfaces"
	"bloodlink/Usecase"
	"context"
	"fmt"
	"log"
	"time"
)

func StartExpirationJob(
	inventoryUC *Usecase.BloodInventoryUsecase,
	bloodReqRepo Interfaces.IBloodRequestRepository,
	notifUC Interfaces.INotificationUsecase,
	donorBloodReqUC *Usecase.DonorBloodRequestUsecase,
	userUC Interfaces.IUserUseCase,
	emergencyUC Interfaces.IEmergencyRequestUsecase,
	campaignRepo Interfaces.ICampaignRepository,
) {
	for {
		log.Println("[JOB] Running expiration and reservation cleanup...")

		// 1. Mark units past expiry date as EXPIRED
		if err := inventoryUC.UpdateExpiredUnits(); err != nil {
			log.Printf("[JOB ERROR] Failed to update expired units: %v", err)
		}

		// 2. Cleanup stale reservations (24h rule)
		requestIDs, err := inventoryUC.ExpireReservations() // inventoryUC now has internal repo access
		if err != nil {
			log.Printf("[JOB ERROR] Failed to expire stale reservations: %v", err)
		}

		// 2.1. Cleanup stale DONOR reservations (24h rule)
		if err := donorBloodReqUC.ExpireStaleRequests(); err != nil {
			log.Printf("[JOB ERROR] Failed to expire stale donor requests: %v", err)
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
				go func(r *Domain.BloodRequest) {
					subject := "Blood Request Auto-Rejected"
					content := fmt.Sprintf("Your request for %s blood has been automatically rejected because the reserved units were not collected within the required 24-hour window.", r.BloodType)
					_ = notifUC.SendToHospital(r.HospitalID, "REJECTED", subject, content)
				}(req)
			}
		}

		// 4. Notify donors who become eligible today (90-day mark)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		if err := userUC.NotifyEligibleDonors(ctx); err != nil {
			log.Printf("[JOB ERROR] Failed to notify eligible donors: %v", err)
		}
		cancel()

		// 5. Auto-complete expired emergencies
		if err := emergencyUC.MarkCompletedEmergencies(); err != nil {
			log.Printf("[JOB ERROR] Failed to mark completed emergencies: %v", err)
		}

		// 6. Auto-close expired campaigns
		if err := campaignRepo.MarkClosedCampaigns(); err != nil {
			log.Printf("[JOB ERROR] Failed to mark closed campaigns: %v", err)
		}

		time.Sleep(1 * time.Hour) // Check more frequently than 24h to be accurate
	}
}