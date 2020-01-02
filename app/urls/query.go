package urls

import (
	"fmt"
	"log"
	"database/sql"
	"strings"

	"github.com/lib/pq"
)

type IUrlsRepository interface {
	GetAllUrls() ([]UrlPair, error)
	CreateUrl(short, long string) error
	GetUserUrls(accountID, userID int64, filters ...LinkFilter) ([]UrlPair, error)
	GetUserUrlsCount(accountID int64) (int, error)
}

type UrlsRepository struct {
	DB     *sql.DB
	Logger *log.Logger
}

func (repo *UrlsRepository) GetAllUrls() ([]UrlPair, error) {

	rows, err := repo.DB.Query("select short_url, full_url from urls where account_id is null")
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

type LinkFilter struct {
	ShortUrl    []string
	FullUrl     []string
	Tags        []string
	FullText    string
}

func (repo *UrlsRepository) GetUserUrls(accountID, userID int64, filters ...LinkFilter) ([]UrlPair, error) {

	query := `
	with url_group as (
		select distinct(ug.url_id) as url_id from urls_groups ug where ug.group_id IN (
			select group_id from users_groups where users_groups.user_id = $2
		)
	), url_tags as (
		select link_id, array_agg(tags.tag) tags_list from tags 
		group by link_id
	)
	select u.short_url, u.full_url, u.tl
	from (
		select *, t.tags_list tl from urls
		left join url_group ug on ug.url_id = urls.id
		left outer join url_tags t on t.link_id = urls.id
		where (urls.account_id = $1 and not exists (select 1 from url_group)) 
		or (ug.url_id is not null and exists (select 1 from url_group))
	) u
	`

	queryArgs := []interface{}{accountID, userID}

	var filterExpressions []string

	for _, f := range filters {
		if len(f.Tags) > 0 {
			exp := []string{fmt.Sprintf("u.tl && $%d", len(queryArgs) + 1)}
			queryArgs = append(queryArgs, pq.Array(f.Tags))
			filterExpressions = append(filterExpressions, fmt.Sprintf("(%s)", strings.Join(exp, " OR ")))
		}
		if len(f.ShortUrl) > 0 {
			exp := []string{}
			for _, v := range f.ShortUrl {
				exp = append(exp, fmt.Sprintf("u.short_url LIKE $%d", len(queryArgs) + 1))
				queryArgs = append(queryArgs, v + "%")
			}
			filterExpressions = append(filterExpressions, fmt.Sprintf("(%s)", strings.Join(exp, " OR ")))
		}
		if len(f.FullUrl) > 0 {
			exp := []string{}
			for _, v := range f.FullUrl {
				exp = append(exp, fmt.Sprintf("u.full_url LIKE $%d", len(queryArgs) + 1))
				queryArgs = append(queryArgs, v + "%")
			}
			filterExpressions = append(filterExpressions, fmt.Sprintf("(%s)", strings.Join(exp, " OR ")))
		}
		if f.FullText != "" {
			exp := []string{
				fmt.Sprintf("u.tl && $%d", len(queryArgs) + 1),
				fmt.Sprintf("u.short_url LIKE $%d", len(queryArgs) + 2),
				fmt.Sprintf("u.full_url LIKE $%d", len(queryArgs) + 3),
			}
			queryArgs = append(queryArgs, pq.Array([]string{f.FullText}))
			queryArgs = append(queryArgs, f.FullText + "%")
			queryArgs = append(queryArgs, f.FullText + "%")
			filterExpressions = append(filterExpressions, fmt.Sprintf("(%s)", strings.Join(exp, " OR ")))
		}
	}

	if len(filterExpressions) > 0 {
		query += "where " + strings.Join(filterExpressions, " AND ")
	}

	rows, err := repo.DB.Query(query, queryArgs...)

	if err != nil {
		return nil, err
	}

	var urls []UrlPair

	for rows.Next() {
		var tagList []string
		var shortURL, fullURL string
		err := rows.Scan(&shortURL, &fullURL, pq.Array(&tagList))
		if err != nil {
			return nil, err
		}
		urls = append(urls, UrlPair{Short: shortURL, Long: fullURL, Tags: tagList})
	}

	return urls, nil
}

func (repo *UrlsRepository) GetUserUrlsCount(accountID int64) (int, error) {

	var count int
	err := repo.DB.QueryRow("select count(*) from urls where account_id = $1", accountID).Scan(&count)
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
