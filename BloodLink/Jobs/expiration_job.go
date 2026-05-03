package Jobs

import (
	"time"
	"bloodlink/Usecase"
)

func StartExpirationJob(u *Usecase.BloodInventoryUsecase) {
	for {
		u.UpdateExpiredUnits()
		time.Sleep(24 * time.Hour)
	}
}