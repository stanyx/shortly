package billing

type BillingPlan struct {
	ID          int64
	Name        string
	Description string
	PeriodType  string
	Price 	    string
	Options     []BillingOption
	AccountID   int64
	IsAnnual    bool
}

type BillingOption struct {
	ID          int64
	Name        string
	Description string
	Value       string
	PlanID      int64
}