package billing

import (

	"database/sql"
)

type BillingRepository struct {
 	DB *sql.DB
}

func (r *BillingRepository) GetBillingPlanCost(planID int64) (string, error) {

	var cost string
	err := r.DB.QueryRow(`
		SELECT bp.price FROM billing_plans bp WHERE bp.id = $1
	`, planID).Scan(&cost)

	if err != nil {
		return "", err
	}

	return cost, nil
}

func (r *BillingRepository) ApplyBillingPlan(accountID, planID int64) error {

	if _, err := r.DB.Exec("UPDATE billing_accounts SET active = false WHERE account_id = $1", accountID); err != nil {
		return err
	}

	if _, err := r.DB.Exec("INSERT INTO billing_accounts (account_id, plan_id, active) VALUES ($1, $2, true)", accountID, planID); err != nil {
		return err
	}

	return nil
}

func (r *BillingRepository) GetBillingPlanOptions(accountID, planID int64) ([]BillingOption, error) {

	rows, err := r.DB.Query(`
		SELECT bp.id, opts.id, opts.name, opts.description, opts.value 
		FROM billing_accounts ubp
		INNER JOIN billing_plans bp ON ubp.plan_id = bp.id
		INNER JOIN billing_options opts ON bp.id = opts.plan_id
		WHERE ubp.account_id = $1 AND ubp.plan_id = $2
		AND ubp.active = true
	`, accountID, planID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var options []BillingOption
	for rows.Next() {
		var bo BillingOption
		if err := rows.Scan(&bo.PlanID, &bo.ID, &bo.Name, &bo.Description, &bo.Value); err != nil {
			return nil, err
		}
		options = append(options, bo)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return options, err
}

func (r *BillingRepository) GetAllBillingPlans() ([]BillingPlan, error) {

	rows1, err := r.DB.Query(`
		SELECT bp.id, opts.id, opts.name, opts.description, opts.value FROM billing_options opts
		LEFT OUTER JOIN billing_plans bp ON bp.id = opts.plan_id
	`)

	if err != nil {
		return nil, err
	}

	defer rows1.Close()

	optionByPlan := make(map[int64][]BillingOption)

	for rows1.Next() {
		var bp BillingOption
		if err := rows1.Scan(&bp.PlanID, &bp.ID, &bp.Name, &bp.Description, &bp.Value); err != nil {
			return nil, err
		}
		optionByPlan[bp.PlanID] = append(optionByPlan[bp.PlanID], bp)
	}

	if err := rows1.Err(); err != nil {
		return nil, err
	}

	rows2, err := r.DB.Query(`
		SELECT id, name, description, price FROM billing_plans
	`)

	if err != nil {
		return nil, err
	}

	defer rows2.Close()

	var plans []BillingPlan
	for rows2.Next() {
		var bp BillingPlan
		if err := rows2.Scan(&bp.ID, &bp.Name, &bp.Description, &bp.Price); err != nil {
			return nil, err
		}
		bp.Options = optionByPlan[bp.ID]
		plans = append(plans, bp)
	}

	if err := rows2.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

func (r *BillingRepository) GetAllUserBillingPlans() ([]BillingPlan, error) {

	rows1, err := r.DB.Query(`
		SELECT ubp.account_id, opts.id, opts.name, opts.description, opts.value 
		FROM billing_accounts ubp
		INNER JOIN billing_options opts ON opts.plan_id = ubp.plan_id
		INNER JOIN billing_plans bp ON bp.id = opts.plan_id
		WHERE ubp.active = true
	`)

	if err != nil {
		return nil, err
	}

	defer rows1.Close()

	optionByAccount := make(map[int64][]BillingOption)

	for rows1.Next() {
		var accountID int64
		var bp BillingOption
		if err := rows1.Scan(&accountID, &bp.ID, &bp.Name, &bp.Description, &bp.Value); err != nil {
			return nil, err
		}
		optionByAccount[accountID] = append(optionByAccount[accountID], bp)
	}

	if err := rows1.Err(); err != nil {
		return nil, err
	}

	rows2, err := r.DB.Query(`
		SELECT ubp.account_id, bp.id, bp.name, bp.description, bp.price 
		FROM billing_accounts ubp
		INNER JOIN billing_plans bp ON bp.id = ubp.plan_id
		WHERE ubp.active = true
	`)

	if err != nil {
		return nil, err
	}

	defer rows2.Close()

	var plans []BillingPlan
	for rows2.Next() {
		var bp BillingPlan
		if err := rows2.Scan(&bp.AccountID, &bp.ID, &bp.Name, &bp.Description, &bp.Price); err != nil {
			return nil, err
		}
		bp.Options = optionByAccount[bp.AccountID]
		plans = append(plans, bp)
	}

	if err := rows2.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}