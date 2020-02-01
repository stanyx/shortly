package billing

import (
	"strconv"
	"time"
)

type BillingPlan struct {
	ID          int64
	Name        string
	Description string
	PeriodType  string
	Price       string
	AnnualPrice string
	Options     []BillingOption
	AccountID   int64
}

type AccountBillingPlan struct {
	BillingPlan
	Start    time.Time
	End      time.Time
	Charge   int
	IsAnnual bool
}

type BillingPlanActivation struct {
	PlanID   int64
	Start    time.Time
	End      time.Time
	Charge   int
	IsAnnual bool
}

type BillingOption struct {
	ID          int64
	Name        string
	Description string
	Value       string
	PlanID      int64
}

func (opt BillingOption) AsInt64() int64 {
	v, _ := strconv.ParseInt(opt.Value, 0, 64)
	return v
}

type BillingParameter struct {
	BillingOption
	CurrentValue string
}
