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

func (r *BillingRepository) GetAllBillingPlans() ([]BillingPlan, error) {

	rows, err := r.DB.Query(`
		SELECT bp.id, opts.id, opts.name, opts.description, opts.value FROM billing_options opts
		LEFT OUTER JOIN billing_plans bp ON bp.id = opts.plan_id
	`)

	if err != nil {
		return nil, err
	}

	optionByPlan := make(map[int64][]BillingOption)

	for rows.Next() {
		var bp BillingOption
		if err := rows.Scan(&bp.PlanID, &bp.ID, &bp.Name, &bp.Description, &bp.Value); err != nil {
			return nil, err
		}
		optionByPlan[bp.PlanID] = append(optionByPlan[bp.PlanID], bp)
	}

	rows, err = r.DB.Query(`
		SELECT id, name, description, price FROM billing_plans
	`)

	if err != nil {
		return nil, err
	}

	var plans []BillingPlan
	for rows.Next() {
		var bp BillingPlan
		if err := rows.Scan(&bp.ID, &bp.Name, &bp.Description, &bp.Price); err != nil {
			return nil, err
		}
		bp.Options = optionByPlan[bp.ID]
		plans = append(plans, bp)
	}

	return plans, nil
}