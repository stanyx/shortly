package billing

import (
	"time"
)

type BillingPlan struct {
	PeriodType string
	Price 	   string
}

type UserBilling struct {
	PlanID int64
	UserID int64
}

type BillingOption struct {
	Name        string
	Description string
	Value       string
	PlanID      int64
}

type UserBillingOption struct {
	UserBillingID   int64
	BillingOptionID int64
}

type BillingPlanHistory struct {
	UserID    int64
	StartedAt time.Time
	StoppedAt time.Time
}