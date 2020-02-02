package dashboards

// Dashboard ...
type Dashboard struct {
	ID          int64
	AccountID   int64
	Name        string
	Description string
	Width       int
	Height      int
	Widgets     []DashboardWidget
}

const (
	ChartWidget   = "chart"
	CounterWidget = "counter"
)

// DashboardWidget ...
type DashboardWidget struct {
	ID      int64
	Title   string
	Type    string
	DataURL string
	PosX    int
	PosY    int
	Span    int
}
