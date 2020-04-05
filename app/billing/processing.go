package billing

import (
	"log"
	"shortly/utils"
	"sync"
	"time"
)

// CheckSubscription ...
func CheckSubscription(acc AccountBillingPlan) error {
	// TODO verify stripe subscriptions
	return nil
}

// SubscriptionCheck ...
func SubscriptionCheck(repo *BillingRepository, billingLimiter *BillingLimiter, logger *log.Logger) {

	go func() {

		for {
			accountPlans, err := repo.GetActiveBillingPlans(0)
			if err != nil {
				logger.Println("accounts fetch error", err)
			}

			var wg sync.WaitGroup

			for _, acc := range accountPlans {
				if acc.End.Before(utils.Now()) {
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

func SyncSubscriptionInformation() {

	go func() {

		for {

			time.Sleep(time.Second)
		}

	}()
}
