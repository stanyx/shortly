package campaigns

import (
	"database/sql"
	"log"
	"time"

	"shortly/app/data"
	"shortly/app/links"
)

// CampaignLink ...
type CampaignLink struct {
	ID          int64
	CampaignID  int64
	ShortUrl    string
	LongUrl     string
	Description string
}

// Repository ...
type Repository struct {
	DB        *sql.DB
	HistoryDB *data.HistoryDB
	Logger    *log.Logger
}

func (r *Repository) GetUserCampaigns(accountID int64) ([]Campaign, error) {

	query := `select cmp.id, l.id, l.short_url, l.long_url, l.description
		from campaigns cmp
		inner join campaigns_links cmpl on cmpl.campaign_id = cmp.id 
		inner join links l on l.id = cmpl.link_id
		where cmp.account_id = $1
	`

	rows, err := r.DB.Query(query, accountID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	linksByCampaign := make(map[int64][]CampaignLink)

	for rows.Next() {
		var link CampaignLink
		err := rows.Scan(&link.CampaignID, &link.ID, &link.ShortUrl, &link.LongUrl, &link.Description)
		if err != nil {
			return nil, err
		}
		linksByCampaign[link.CampaignID] = append(linksByCampaign[link.CampaignID], link)
	}

	campaignsQuery := `select id, name, description from campaigns where account_id = $1`
	campaignRows, err := r.DB.Query(campaignsQuery, accountID)

	if err != nil {
		return nil, err
	}

	var list []Campaign

	for campaignRows.Next() {
		var cmp Campaign
		err := campaignRows.Scan(&cmp.ID, &cmp.Name, &cmp.Description)
		if err != nil {
			return nil, err
		}

		for _, l := range linksByCampaign[cmp.ID] {
			cmp.Links = append(cmp.Links, links.Link{
				ID:          l.ID,
				Short:       l.ShortUrl,
				Long:        l.LongUrl,
				Description: l.Description,
			})
		}

		list = append(list, cmp)
	}

	defer campaignRows.Close()

	if err := campaignRows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

// CampaignClickData ...
type CampaignClickData struct {
	ShortURL string
	Data     []data.CounterData
}

func (r *Repository) GetCampaignClicksData(accountID, campaignID int64, startTime, endTime time.Time) ([]CampaignClickData, error) {

	query := `select l.id, l.short_url
		from campaigns cmp
		inner join campaigns_links cmpl on cmpl.campaign_id = cmp.id 
		inner join links l on l.id = cmpl.link_id
		where cmp.account_id = $1 and cmpl.campaign_id = $2
	`

	rows, err := r.DB.Query(query, accountID, campaignID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	var list []CampaignClickData

	for rows.Next() {
		var linkID int64
		var shortURL string
		if err := rows.Scan(&linkID, &shortURL); err != nil {
			return nil, err
		}
		data, err := r.HistoryDB.GetClicksData(accountID, shortURL, startTime, endTime)
		if err != nil {
			return nil, err
		}
		list = append(list, CampaignClickData{
			ShortURL: shortURL,
			Data:     data,
		})
	}

	return list, nil
}

func (r *Repository) CreateCampaign(cmp Campaign) (int64, error) {
	var rowID int64

	err := r.DB.QueryRow(`
		insert into "campaigns" (name, description, account_id) 
		values ($1, $2, $3) returning id`,
		cmp.Name, cmp.Description, cmp.AccountID,
	).Scan(&rowID)

	if err != nil {
		return 0, err
	}

	return rowID, nil
}

func (r *Repository) StartCampaign(cmpID int64) error {
	_, err := r.DB.Exec(`update "campaigns" set active = true where id = $1`, cmpID)
	return err
}

func (r *Repository) StopCampaign(cmpID int64) error {
	_, err := r.DB.Exec(`update "campaigns" set active = false where id = $1`, cmpID)
	return err
}

func (r *Repository) DeleteCampaign(cmpID int64) error {
	_, err := r.DB.Exec(`delete from "campaigns" where id = $1`, cmpID)
	return err
}

func (r *Repository) AddLinkToCampaign(cmpID, linkID int64, utm UTMSetting) (int64, error) {
	var rowID int64
	err := r.DB.QueryRow(`
		insert into "campaigns_links" (campaign_id, link_id, utm_source, utm_medium, utm_term, utm_content) 
		values ($1, $2, $3, $4, $5, $6)
		returning id
	`, cmpID, linkID, utm.Source, utm.Medium, utm.Term, utm.Content).Scan(&rowID)
	return rowID, err
}

func (r *Repository) DeleteLinkFromCampaign(cmpID, linkID int64) error {
	_, err := r.DB.Exec(`delete from "campaigns_links" where campaign_id = $1 and link_id = $2`, cmpID, linkID)
	return err
}
