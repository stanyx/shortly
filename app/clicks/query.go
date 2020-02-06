package clicks

import (
	"database/sql"
	"log"
)

// Repository ...
type Repository struct {
	DB     *sql.DB
	Logger *log.Logger
}

// GetTotalClicks ...
func (r *Repository) GetTotalClicks(accountID int64) (int64, error) {

	var count int64
	if err := r.DB.QueryRow(`
		select count(*) from redirect_log r
		inner join links l on l.short_url = r.short_url
		where l.account_id = $1
	`, accountID).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

// GetClicksData ...
func (r *Repository) GetClicksData(accountID int64) ([]ClickData, error) {

	rows, err := r.DB.Query(`
	select d, count(r.id) 
	from generate_series(date_trunc('day', now() at time zone 'utc'), date_trunc('day', now() at time zone 'utc') + '1 day'::interval, '1 hour'::interval) as d
	left join (
		select * from redirect_log
		where timestamp >= '2020-01-01 00:00:00' and
		timestamp < (date_trunc('day', now() at time zone 'utc') + '1 day'::interval)
	) r on date_trunc('hour', r.timestamp at time zone 'utc') = d
	left join (select * from links where account_id = $1) l on l.short_url = r.short_url
	where d >= '2020-01-01 00:00:00' and
		  d < (date_trunc('day', now() at time zone 'utc') + '1 day'::interval)
	group by d
	`, accountID)

	if err != nil {
		return nil, err
	}

	var list []ClickData
	for rows.Next() {
		var u ClickData
		err := rows.Scan(&u.Time, &u.Count)
		if err != nil {
			return nil, err
		}
		list = append(list, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer rows.Close()

	return list, nil
}

// GetClicksDataByDay ...
func (r *Repository) GetClicksDataByDay(shortURL string) ([]ClickData, error) {

	rows, err := r.DB.Query(`
	select date_trunc('day', timestamp at time zone 'utc') t, count(*) from redirect_log where short_url = $1
	group by t
	`, shortURL)

	if err != nil {
		return nil, err
	}

	var list []ClickData
	for rows.Next() {
		var u ClickData
		err := rows.Scan(&u.Time, &u.Count)
		if err != nil {
			return nil, err
		}
		list = append(list, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer rows.Close()

	return list, nil
}

func (r *Repository) GetLinkInfoByDay(shortURL string) ([]LinkData, error) {

	rows, err := r.DB.Query(`
	select date_trunc('day', timestamp at time zone 'utc') t, 
	country, referer, count(*) from redirect_log where short_url = $1
	group by t, country, referer
	`, shortURL)

	if err != nil {
		return nil, err
	}

	var list []LinkData
	for rows.Next() {
		var u LinkData
		err := rows.Scan(&u.Time, &u.Location, &u.Referer, &u.Count)
		if err != nil {
			return nil, err
		}
		list = append(list, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer rows.Close()

	return list, nil
}
