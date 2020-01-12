package billing

import (
	"database/sql"
)

type BillingRepository struct {
	DB *sql.DB
}

func (r *BillingRepository) GetBillingPlanCost(planID int64, isAnnual bool) (string, error) {

	var cost string

	queryArgs := []interface{}{planID}

	query := `
		SELECT p.price FROM billing_plans bp
		INNER JOIN billing_price p ON bp.id = p.plan_id
		WHERE bp.id = $1
	`

	if isAnnual {
		query += " AND p.is_annual = true"
	} else {
		query += " AND p.is_annual = false"
	}

	err := r.DB.QueryRow(query, queryArgs...).Scan(&cost)
	if err != nil {
		return "", err
	}

	return cost, nil
}

func (r *BillingRepository) ApplyBillingPlan(accountID, planID int64) error {

	if _, err := r.DB.Exec("UPDATE billing_accounts SET active = false WHERE account_id = $1", accountID); err != nil {
		return err
	}

	if _, err := r.DB.Exec(`
		insert into "billing_accounts" (account_id, plan_id, started_at, ended_at, active) 
		values ($1, $2, date_trunc('day', now()), date_trunc('day', now()) + '30 day'::interval, true)
	`, accountID, planID); err != nil {
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
		SELECT a.id, a.name, a.description, a.prices[0] price, a.prices[1] annual_price FROM (
			SELECT bp.id, bp.name, bp.description, array_agg(p.price) prices
			INNER JOIN billing_price p ON p.plan_id = bp.id
			FROM billing_plans bp
			GROUP BY bp.id, bp.name, bp.description, p.is_annual
			ORDER BY p.is_annual
		) a
	`)

	if err != nil {
		return nil, err
	}

	defer rows2.Close()

	var plans []BillingPlan
	for rows2.Next() {
		var bp BillingPlan
		if err := rows2.Scan(&bp.ID, &bp.Name, &bp.Description, &bp.Price, &bp.AnnualPrice); err != nil {
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

func (r *BillingRepository) GetActiveBillingPlans(accountID int64) ([]AccountBillingPlan, error) {
	query := `
		SELECT ubp.account_id, ubp.started_at, ubp.ended_at, bp.id, bp.name, bp.description 
		FROM billing_accounts ubp
		INNER JOIN billing_plans bp ON bp.id = ubp.plan_id
		WHERE ubp.active = true
	`

	var queryArgs []interface{}
	if accountID > 0 {
		query += " AND ubp.account_id = $1"
		queryArgs = append(queryArgs, accountID)
	}
	rows, err := r.DB.Query(query, queryArgs...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var plans []AccountBillingPlan
	for rows.Next() {
		var bp AccountBillingPlan
		if err := rows.Scan(
			&bp.AccountID,
			&bp.Start,
			&bp.End,
			&bp.ID,
			&bp.Name,
			&bp.Description,
		); err != nil {
			return nil, err
		}
		plans = append(plans, bp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return plans, nil
}

func (r *BillingRepository) GetPlansOptions(accountID int64) (map[int64][]BillingOption, error) {

	query := `
		SELECT ubp.account_id, opts.id, opts.name, opts.description, opts.value 
		FROM billing_accounts ubp
		INNER JOIN billing_options opts ON opts.plan_id = ubp.plan_id
		INNER JOIN billing_plans bp ON bp.id = opts.plan_id
		WHERE ubp.active = true
	`

	var queryArgs []interface{}
	if accountID > 0 {
		query += " AND ubp.account_id = $1"
		queryArgs = append(queryArgs, accountID)
	}

	rows, err := r.DB.Query(query, queryArgs...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	optionByAccount := make(map[int64][]BillingOption)

	for rows.Next() {
		var accountID int64
		var bp BillingOption
		if err := rows.Scan(&accountID, &bp.ID, &bp.Name, &bp.Description, &bp.Value); err != nil {
			return nil, err
		}
		optionByAccount[accountID] = append(optionByAccount[accountID], bp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return optionByAccount, nil
}

func (r *BillingRepository) GetAllUserBillingPlans(accountID int64) ([]AccountBillingPlan, error) {

	optionByAccount, err := r.GetPlansOptions(accountID)
	if err != nil {
		return nil, err
	}

	plans, err := r.GetActiveBillingPlans(accountID)
	if err != nil {
		return nil, err
	}

	for _, bp := range plans {
		bp.Options = optionByAccount[bp.AccountID]
	}

	return plans, nil
}
