package billing

import (
	"log"
	"sync"
	"time"
)

// CheckSubscription ...
func CheckSubscription(acc AccountBillingPlan) error {
	// TODO verify stripe subscriptions
	return nil
}

// SubsriptionCheck ...
func SubsriptionCheck(repo *BillingRepository, billingLimiter *BillingLimiter, logger *log.Logger) {

	go func() {

		for {
			accountPlans, err := repo.GetActiveBillingPlans(0)
			if err != nil {
				logger.Println("accounts fetch error", err)
			}

			var wg sync.WaitGroup

			for _, acc := range accountPlans {
				if acc.End.Before(time.Now()) {
					wg.Add(1)
					go func(acc AccountBillingPlan) {
						lock := billingLimiter.Lock(acc.ID)
						err := CheckSubscription(acc)
						if err != nil {
							logger.Println("subscription check error", err)
							if err := billingLimiter.DowngradeToDefaultPlan(acc.ID); err != nil {
								logger.Println("downgrade to defaul plan error", err)
							}
						}
						lock.Unlock()
						wg.Done()
					}(acc)
				}
			}

			wg.Wait()

			time.Sleep(time.Second)
		}

	}()

}
