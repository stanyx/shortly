package billing

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/customer"
	"github.com/stripe/stripe-go/sub"
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

func (r *BillingRepository) ApplyBillingPlan(accountID int64, activation BillingPlanActivation) error {

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec("UPDATE billing_accounts SET active = false WHERE account_id = $1", accountID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if activation.Start.Unix() == 0 {
		activation.Start = time.Now()
	}

	if activation.End.Unix() == 0 {
		activation.End = time.Now().Add(time.Hour * 24 * 30)
	}

	if _, err := tx.Exec(`
		insert into "billing_accounts" (account_id, plan_id, started_at, ended_at, charge, is_annual, active) 
		values ($1, $2, $3, $4, $5, $6, true)
	`, accountID, activation.PlanID, activation.Start, activation.End, activation.Charge, activation.IsAnnual); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (r *BillingRepository) GetDefaultPlanOptions(planID int64) ([]BillingOption, error) {
	rows, err := r.DB.Query(`
		SELECT bp.id, opts.id, opts.name, opts.description, opts.value 
		FROM billing_plans bp
		INNER JOIN billing_options opts ON bp.id = opts.plan_id
		WHERE bp.id = $1
	`, planID)

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
		SELECT bp.id, ubp.account_id, ubp.started_at, ubp.ended_at, ubp.charge, ubp.is_annual,
		bp.id, bp.name, bp.description 
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
			&bp.ID,
			&bp.AccountID,
			&bp.Start,
			&bp.End,
			&bp.Charge,
			&bp.IsAnnual,
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

// TODO rename to GetAccountBillingPlans
func (r *BillingRepository) GetAllUserBillingPlans(accountID int64) ([]AccountBillingPlan, error) {

	optionByAccount, err := r.GetPlansOptions(accountID)
	if err != nil {
		return nil, err
	}

	plans, err := r.GetActiveBillingPlans(accountID)
	if err != nil {
		return nil, err
	}

	for i, bp := range plans {
		plans[i].Options = optionByAccount[bp.AccountID]
	}

	return plans, nil
}

func (r *BillingRepository) GetDefaultPlan() (*BillingPlan, error) {

	var defaultPlan BillingPlan

	err := r.DB.QueryRow("select id from billing_plans where name = 'free' limit 1").Scan(&defaultPlan.ID)
	if err != nil {
		return nil, err
	}

	options, err := r.GetDefaultPlanOptions(defaultPlan.ID)
	if err != nil {
		return nil, err
	}

	defaultPlan.Options = options

	return &defaultPlan, nil
}

func (r *BillingRepository) GetBillingPlansToUpgrade(planID int64) ([]BillingPlan, error) {

	rows, err := r.DB.Query(`
		select bi.id, bi.name, prices[1], prices[2] from (
			select bp.id, bp.name, array_agg(bpr.price) prices
			from billing_plans bp
			left join billing_price bpr on bpr.plan_id = bp.id  
			where 
			upgrade_rate > (select upgrade_rate from billing_plans where id = $1)
			group by bp.id, bp.name, bp.upgrade_rate
			order by bp.upgrade_rate
		) bi
	`, planID)

	if err != nil {
		return nil, err
	}

	var list []BillingPlan

	for rows.Next() {
		var bp BillingPlan
		err := rows.Scan(&bp.ID, &bp.Name, &bp.Price, &bp.AnnualPrice)
		if err != nil {
			return nil, err
		}
		list = append(list, bp)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer rows.Close()

	return list, nil

}

func (r *BillingRepository) GetAccountBillingPlan(accountID int64) (*AccountBillingPlan, error) {
	plans, err := r.GetAllUserBillingPlans(accountID)
	if err != nil {
		return nil, err
	}

	if len(plans) == 0 {
		return nil, errors.New("no billing plan found")
	}

	plan := plans[0]

	return &plan, nil
}

func (r *BillingRepository) IsAttachToPlan(accountID int64) (bool, error) {

	plans, err := r.GetAllUserBillingPlans(accountID)
	if err != nil {
		return false, err
	}

	if len(plans) == 0 {
		return false, err
	}

	return true, nil
}

var ErrBillingAccountAlreadyExists = errors.New("account already attached to billing")

func (r *BillingRepository) AttachToDefaultBilling(accountID, planID int64) (*BillingAccount, error) {

	ok, err := r.IsAttachToPlan(accountID)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, ErrBillingAccountAlreadyExists
	}

	tNow := time.Now()
	start := time.Date(tNow.Year(), tNow.Month(), tNow.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Duration(24*365*100) * time.Hour)

	planActivation := BillingPlanActivation{
		PlanID: planID,
		Start:  start,
		End:    end,
	}

	if err := r.ApplyBillingPlan(accountID, planActivation); err != nil {
		return nil, errors.Wrap(err, "billing plan apply error:")
	}

	options, err := r.GetBillingPlanOptions(accountID, planID)
	if err != nil {
		return nil, errors.Wrap(err, "get billing plan options error:")
	}

	billingAccount := &BillingAccount{
		Start:   start,
		End:     end,
		Options: options,
	}

	return billingAccount, err
}

func (r *BillingRepository) GetStripeSubscriptionID(planID int64) (string, error) {
	var stripeID string
	err := r.DB.QueryRow("select stripe_id from billing_plans where id = $1", planID).Scan(&stripeID)
	return stripeID, err
}

func (r *BillingRepository) GetAccountIDByStripeCharge(stripeID string) (int64, error) {
	var accountID int64
	err := r.DB.QueryRow("select account_id from stripe_charges where stripe_id = $1", stripeID).Scan(&accountID)
	return accountID, err
}

func (r *BillingRepository) CancelSubscription(accountID int64) error {

	var subscriptionID string

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	err = tx.QueryRow("update stripe_subscriptions set active = false where account_id = $1 returning stripe_id", accountID).Scan(&subscriptionID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = sub.Cancel(subscriptionID, nil)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// CancelSubscriptionExternal set subscription status to active = false from external event (webhook)
func (r *BillingRepository) CancelSubscriptionExternal(stripeID string, timestamp int64) (int64, error) {
	var accountID int64
	err := r.DB.QueryRow(
		`update stripe_subscriptions set active = false, canceled_at = $1 
		 where stripe_id = $2 returning account_id
		`, timestamp, stripeID).Scan(&accountID)
	return accountID, err
}

func (r *BillingRepository) CreateStripeCustomer(accountID int64, email string) error {

	params := &stripe.CustomerParams{
		Description: stripe.String(fmt.Sprintf("Account<%d>", accountID)),
		Email:       stripe.String(email),
	}

	customer, err := customer.New(params)
	if err != nil {
		return err
	}

	_, err = r.DB.Exec("insert into stripe_customers (account_id, stripe_id, created) values ($1, $2, $3)",
		accountID, customer.ID, customer.Created)

	return err
}

func (r *BillingRepository) GetStripeCustomer(accountID int64) (string, error) {
	var stripeID string
	err := r.DB.QueryRow(
		`select stripe_id from stripe_customers where account_id = $1`, accountID).Scan(&stripeID)
	return stripeID, err
}

func (r *BillingRepository) CreateStripeSubscription(accountID, planID int64, s *stripe.Subscription) error {
	_, err := r.DB.Exec("insert into stripe_subscriptions (account_id, plan_id, stripe_id, created) values ($1, $2, $3, $4)",
		accountID, planID, s.ID, s.Created)
	return err
}

func (r *BillingRepository) CreateStripeCharge(accountID, planID int64, ch *stripe.Charge) error {
	_, err := r.DB.Exec("insert into stripe_charges (account_id, plan_id, stripe_id, created) values ($1, $2, $3, $4)",
		accountID, planID, ch.ID, ch.Created)
	return err
}

func (r *BillingRepository) CreateStripeEvent(id string, ev string, timestamp int64, eventErr error) error {

	var errMessage string
	if eventErr != nil {
		errMessage = eventErr.Error()
	}

	_, err := r.DB.Exec("insert into stripe_events (stripe_id, timestamp, payload, error) values ($1, $2, $3, $4)",
		id, timestamp, ev, errMessage)
	return err
}
