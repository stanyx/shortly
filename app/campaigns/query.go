package campaigns

import (
	"database/sql"
	"log"
	"time"

	"shortly/app/data"
)

// CampaignLink ...
type CampaignLink struct {
	ID          int64
	CampaignID  int64
	ShortUrl    string
	LongUrl     string
	Description string
	ChannelID   int64
	ChannelName string
}

// CampaignRepository ...
type CampaignRepository interface {
	GetUserCampaigns(accountID int64) ([]Campaign, error)
}

// Repository ...
type Repository struct {
	DB        *sql.DB
	HistoryDB *data.HistoryDB
	Logger    *log.Logger
}

// GetUserCampaigns ...
func (r *Repository) GetUserCampaigns(accountID int64) ([]Campaign, error) {

	query := `
		select cmp.id, l.id, l.short_url, l.long_url, l.description, ch.id, chs.name
		from campaigns cmp
		inner join campaigns_channels ch on ch.campaign_id = cmp.id
		inner join channels chs on chs.id = ch.channel_id
		inner join campaigns_channels_links cmpl on cmpl.chan_campaign_id = ch.id 
		inner join links l on l.id = cmpl.link_id
		where cmp.account_id = $1
	`

	rows, err := r.DB.Query(query, accountID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	linksByCampaign := make(map[int64][]CampaignLink)

	for rows.Next() {
		var link CampaignLink
		err := rows.Scan(
			&link.CampaignID,
			&link.ID,
			&link.ShortUrl,
			&link.LongUrl,
			&link.Description,
			&link.ChannelID,
			&link.ChannelName,
		)
		if err != nil {
			return nil, err
		}
		linksByCampaign[link.CampaignID] = append(linksByCampaign[link.CampaignID], link)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	campaignsQuery := `select id, name, description from campaigns where account_id = $1`
	campaignRows, err := r.DB.Query(campaignsQuery, accountID)

	if err != nil {
		return nil, err
	}

	defer campaignRows.Close()

	var list []Campaign

	for campaignRows.Next() {
		var cmp Campaign
		err := campaignRows.Scan(&cmp.ID, &cmp.Name, &cmp.Description)
		if err != nil {
			return nil, err
		}

		cmp.Links = append(cmp.Links, linksByCampaign[cmp.ID]...)

		list = append(list, cmp)
	}

	if err := campaignRows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

// CampaignClickData ...
type CampaignClickData struct {
	ShortURL    string
	ChannelID   int64
	ChannelName string
	Data        []data.CounterData
}

// GetCampaignClicksData ...
func (r *Repository) GetCampaignClicksData(accountID, campaignID int64, startTime, endTime time.Time) ([]CampaignClickData, error) {

	query := `select l.id, l.short_url, cl.channel_id, chs.name
		from campaigns cmp
		inner join campaigns_channels ch on ch.campaign_id = cmp.id
		inner join channels chs on chs.id = ch.channel_id
		inner join campaigns_channels_links cl on cl.chan_campaign_id = ch.id
		inner join links l on l.id = cl.link_id
		where cmp.account_id = $1 and ch.campaign_id = $2
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
			Data:     data.Clicks,
		})
	}

	return list, nil
}

// CreateCampaign ...
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

// StartCampaign ...
func (r *Repository) StartCampaign(cmpID int64) error {
	_, err := r.DB.Exec(`update "campaigns" set active = true where id = $1`, cmpID)
	return err
}

// StopCampaign ...
func (r *Repository) StopCampaign(cmpID int64) error {
	_, err := r.DB.Exec(`update "campaigns" set active = false where id = $1`, cmpID)
	return err
}

// DeleteCampaign ...
func (r *Repository) DeleteCampaign(cmpID int64) error {
	_, err := r.DB.Exec(`delete from "campaigns" where id = $1`, cmpID)
	return err
}

// AddChannelToCampaign ...
func (r *Repository) AddChannelsToCampaign(cmpID int64, chIDs []int64) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	for _, chID := range chIDs {
		_, err = tx.Exec(`
			insert into "campaigns_channels" (campaign_id, channel_id) 
			values ( $1, $2 )
		`, cmpID, chID)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

// DeleteChannelToCampaign ...
func (r *Repository) DeleteChannelFromCampaign(cmpID int64, chID int64) (int64, error) {
	var rowID int64
	err := r.DB.QueryRow(`
		delete from "campaigns_channels" where campaign_id = $1 and channel_id = $2
		returning id
	`, cmpID, chID).Scan(&rowID)
	return rowID, err
}

// AddLinkToCampaignChannel ...
func (r *Repository) AddLinkToCampaignChannel(cmpID, channelID, linkID int64, utm UTMSetting) (int64, error) {
	var rowID int64
	err := r.DB.QueryRow(`
		insert into "campaigns_channels_links" (chan_campaign_id, link_id, utm_source, utm_medium, utm_term, utm_content) 
		values ((select id from campaigns_channels where campaign_id = $1 and channel_id = $2 limit 1), $3, $4, $5, $6, $7)
		returning id
	`, cmpID, channelID, linkID, utm.Source, utm.Medium, utm.Term, utm.Content).Scan(&rowID)
	return rowID, err
}

// DeleteLinkFromCampaignChannel ...
func (r *Repository) DeleteLinkFromCampaignChannel(cmpID, channelID, linkID int64) error {
	_, err := r.DB.Exec(`
	delete from "campaigns_channels_links" 
	where chan_campaign_id in (
		select id from campaigns_channels 
		where campaign_id = $1 and channel_id = $2 limit 1
	) and link_id = $3
	`, cmpID, channelID, linkID)
	return err
}

// GetChannels ...
func (r *Repository) GetChannels(accountID, campaignID int64) ([]Channel, error) {

	query := `
		select c.id, c.name from channels c
		left join campaigns_channels ch on ch.channel_id = c.id and ch.campaign_id = $1
		where c.account_id = $2 and ch.id is null
	`
	rows, err := r.DB.Query(query, campaignID, accountID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	list := make([]Channel, 0)

	for rows.Next() {
		var row Channel
		err := rows.Scan(&row.ID, &row.Name)
		if err != nil {
			return nil, err
		}
		list = append(list, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

// CreateChannel ...
func (r *Repository) CreateChannel(accountID int64, row *Channel) (int64, error) {
	var rowID int64
	query := "insert into channels(account_id, name) values ($1, $2) returning id"
	err := r.DB.QueryRow(query, accountID, row.Name).Scan(&rowID)
	return rowID, err
}

// GetCampaignChannels ...
func (r *Repository) GetCampaignChannels(accountID, cmpID int64) ([]Channel, error) {

	query := `
		select c.id, c.name from campaigns_channels ch
		inner join channels c on c.id = ch.channel_id
		where c.account_id = $1 and ch.campaign_id = $2
	`
	rows, err := r.DB.Query(query, accountID, cmpID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	list := make([]Channel, 0)

	for rows.Next() {
		var row Channel
		err := rows.Scan(&row.ID, &row.Name)
		if err != nil {
			return nil, err
		}
		list = append(list, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

// GetChannelLinks ...
func (r *Repository) GetChannelLinks(accountID, campaingID, channelID int64) ([]Link, error) {

	query := `
		select l.id, l.short_url, l.long_url from links l
		left join campaigns_channels_links cmpl on l.id = cmpl.link_id
		left join campaigns_channels ch on ch.id = cmpl.chan_campaign_id
		where l.account_id = $1 and ch.campaign_id = $2 and ch.channel_id = $3
	`
	rows, err := r.DB.Query(query, accountID, campaingID, channelID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	list := make([]Link, 0)

	for rows.Next() {
		var row Link
		err := rows.Scan(&row.ID, &row.Short, &row.Long)
		if err != nil {
			return nil, err
		}
		list = append(list, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

// Link ...
type Link struct {
	ID    int64
	Short string
	Long  string
}

// GetCampaignFreeLinks ...
func (r *Repository) GetCampaignFreeLinks(accountID, channelID int64) ([]Link, error) {

	query := `
		select l.id, l.short_url, l.long_url from links l
		left join campaigns_channels_links cmpl on l.id = cmpl.link_id
		where l.account_id = $1 and cmpl.id is null
	`
	rows, err := r.DB.Query(query, accountID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	list := make([]Link, 0)

	for rows.Next() {
		var row Link
		err := rows.Scan(&row.ID, &row.Short, &row.Long)
		if err != nil {
			return nil, err
		}
		list = append(list, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}
