package dashboards

import (
	"database/sql"
	"log"

	"shortly/app/data"
)

type Repository struct {
	DB        *sql.DB
	HistoryDB *data.HistoryDB
	Logger    *log.Logger
}

func (r *Repository) GetDashboards(accountID int64) ([]Dashboard, error) {
	query := `select id, name, description, width, height from "dashboards"
		where account_id = $1
	`
	rows, err := r.DB.Query(query, accountID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var list []Dashboard
	for rows.Next() {
		var widget Dashboard
		if err := rows.Scan(&widget.ID, &widget.Name, &widget.Description, &widget.Width, &widget.Height); err != nil {
			return nil, err
		}
		list = append(list, widget)
	}

	return list, err
}

func (r *Repository) GetDashboardWidgets(accountID, dashboardID int64) ([]DashboardWidget, error) {
	query := `select id, title, widget_type, data_url, x, y, span from "dashboards_widgets"
		where account_id = $1 and dashboard_id = $2
	`
	rows, err := r.DB.Query(query, accountID, dashboardID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var widgets []DashboardWidget
	for rows.Next() {
		var widget DashboardWidget
		if err := rows.Scan(&widget.ID, &widget.Title, &widget.Type, &widget.DataURL, &widget.PosX, &widget.PosY, &widget.Span); err != nil {
			return nil, err
		}
		widgets = append(widgets, widget)
	}

	return widgets, err
}

func (r *Repository) CreateDashboard(accountID int64, d Dashboard) (int64, error) {
	var rowID int64

	err := r.DB.QueryRow(`
		insert into "dashboards" (account_id, name, description, width, height) 
		values ($1, $2, $3, $4, $5) returning id`,
		accountID, d.Name, d.Description, d.Width, d.Height,
	).Scan(&rowID)

	if err != nil {
		return 0, err
	}

	return rowID, nil
}

func (r *Repository) AddWidget(accountID, dashboardID int64, widget DashboardWidget) error {
	_, err := r.DB.Exec(`
		insert into "dashboards_widgets" (account_id, dashboard_id, title, widget_type, data_url, x, y, span)
		values ( $1, $2, $3, $4, $5, $6, $7, $8 ) returning id
	`, accountID, dashboardID, widget.Title, widget.Type, widget.DataURL, widget.PosX, widget.PosY, widget.Span)
	return err
}

func (r *Repository) DeleteDashboard(accountID, dashboardID int64) error {
	_, err := r.DB.Exec(`
		delete from "dashboards" where id = $1 and account_id = $2
	`, dashboardID, accountID)
	return err
}

func (r *Repository) DeleteDashboardWidget(accountID, dashboardID, widgetID int64) error {
	_, err := r.DB.Exec(`
		delete from "dashboards_widgets" where id = $1 and dashboard_id = $2 and account_id = $3
	`, widgetID, dashboardID, accountID)
	return err
}
