package urls

import (
	"log"
	"database/sql"
)

type IUrlsRepository interface {
	GetAllUrls() ([]UrlPair, error)
	CreateUrl(short, long string) error
	GetUserUrls(accountID, userID int64) ([]UrlPair, error)
	GetUserUrlsCount(userID int64) (int, error)
}

type UrlsRepository struct {
	DB     *sql.DB
	Logger *log.Logger
}

func (repo *UrlsRepository) GetAllUrls() ([]UrlPair, error) {

	rows, err := repo.DB.Query("select short_url, full_url from urls where user_id is null")
	if err != nil {
		return nil, err
	}

	var urls []UrlPair

	for rows.Next() {
		var shortURL, fullURL string
		err := rows.Scan(&shortURL, &fullURL)
		if err != nil {
			return nil, err
		}
		urls = append(urls, UrlPair{Short: shortURL, Long: fullURL})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer rows.Close()

	return urls, nil
}

func (repo *UrlsRepository) CreateUrl(short, long string) error {
	_, err := repo.DB.Exec("INSERT INTO urls (short_url, full_url) VALUES ($1, $2)", short, long)
	return err
}

func (repo *UrlsRepository) GetUserUrls(accountID, userID int64) ([]UrlPair, error) {

	rows, err := repo.DB.Query(`
	with url_group as (
		select distinct(ug.url_id) as url_id from urls_groups ug where ug.group_id IN (
			select group_id from users_groups where users_groups.user_id = $2
		)
	)
	select u.short_url, u.full_url 
	from (
		select * from urls
		left join url_group ug on ug.url_id = urls.id
		where (urls.user_id = $1 and not exists (select 1 from url_group)) 
		or (ug.url_id is not null and exists (select 1 from url_group))
	) u
	`, accountID, userID)

	if err != nil {
		return nil, err
	}

	var urls []UrlPair

	for rows.Next() {
		var shortURL, fullURL string
		err := rows.Scan(&shortURL, &fullURL)
		if err != nil {
			return nil, err
		}
		urls = append(urls, UrlPair{Short: shortURL, Long: fullURL})
	}

	return urls, nil
}

func (repo *UrlsRepository) GetUserUrlsCount(userID int64) (int, error) {

	var count int
	err := repo.DB.QueryRow("select count(*) from urls where user_id = $1", userID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repo *UrlsRepository) AddUrlToGroup(groupID, urlID int64) error {
	_, err := repo.DB.Exec(`
		insert into urls_groups (group_id, url_id) values ($1, $2)
	`, groupID, urlID)
	return err
}

func (repo *UrlsRepository) DeleteUrlFromGroup(groupID, urlID int64) error {
	_, err := repo.DB.Exec(`
		delete from urls_groups where group_id = $1 and url_id = $2
	`, groupID, urlID)
	return err
}
