package billing

import (
	"database/sql"
)

type BillingRepository struct {
 	DB *sql.DB
}

func (r *BillingRepository) ApplyBillingPlan(userID, planID int64) error {
	if _, err := r.DB.Exec("INSERT INTO user_billing VALUES (?, ?)", userID, planID); err != nil {
		return err
	}
	return nil
}