package links

import (
	"fmt"
	"log"
	"database/sql"
	"strings"

	"github.com/lib/pq"
)

type ILinksRepository interface {
	GetAllLinks() ([]Link, error)
	CreateLink(*Link) error
	CreateUserLink(accountID int64, link *Link) (int64, error)
	DeleteUserLink(accountID int64, shortURL string) (int64, error)
	GetUserLinks(accountID, userID int64, filters ...LinkFilter) ([]Link, error)
	GetUserLinksCount(accountID int64) (int, error)
}

type LinksRepository struct {
	DB     *sql.DB
	Logger *log.Logger
}

func (repo *LinksRepository) GetAllLinks() ([]Link, error) {

	rows, err := repo.DB.Query("select short_url, long_url from links where account_id is null")
	if err != nil {
		return nil, err
	}

	var list []Link

	for rows.Next() {
		var shortURL, longURL string
		err := rows.Scan(&shortURL, &longURL)
		if err != nil {
			return nil, err
		}
		list = append(list, Link{
			Short: shortURL, 
			Long: longURL,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	defer rows.Close()

	return list, nil
}

func (repo *LinksRepository) CreateLink(link *Link) error {
	_, err := repo.DB.Exec(`
		insert into "links" (short_url, long_url, description) VALUES ( $1, $2, $3 )
	`, link.Short, link.Long, link.Description)
	return err
}

type LinkFilter struct {
	ShortUrl    []string
	LongUrl     []string
	Tags        []string
	FullText    string
}

func (repo *LinksRepository) GetUserLinks(accountID, userID int64, filters ...LinkFilter) ([]Link, error) {

	query := `
	with url_group as (
		select distinct(ug.link_id) as link_id from links_groups ug where ug.group_id IN (
			select group_id from users_groups where users_groups.user_id = $2
		)
	), url_tags as (
		select link_id, array_agg(tags.tag) tags_list from tags 
		group by link_id
	)
	select u.short_url, u.long_url, u.description, u.tl
	from (
		select *, t.tags_list tl from links
		left join url_group ug on ug.link_id = links.id
		left outer join url_tags t on t.link_id = links.id
		where (links.account_id = $1 and not exists (select 1 from url_group)) 
		or (ug.link_id is not null and exists (select 1 from url_group))
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
		if len(f.LongUrl) > 0 {
			exp := []string{}
			for _, v := range f.LongUrl {
				exp = append(exp, fmt.Sprintf("u.long_url LIKE $%d", len(queryArgs) + 1))
				queryArgs = append(queryArgs, v + "%")
			}
			filterExpressions = append(filterExpressions, fmt.Sprintf("(%s)", strings.Join(exp, " OR ")))
		}
		if f.FullText != "" {
			exp := []string{
				fmt.Sprintf("u.tl && $%d", len(queryArgs) + 1),
				fmt.Sprintf("u.short_url LIKE $%d", len(queryArgs) + 2),
				fmt.Sprintf("u.long_url LIKE $%d", len(queryArgs) + 3),
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

	var list []Link

	for rows.Next() {
		var link Link
		err := rows.Scan(&link.Short, &link.Long, &link.Description, pq.Array(&link.Tags))
		if err != nil {
			return nil, err
		}
		list = append(list, link)
	}

	return list, nil
}

func (repo *LinksRepository) GetUserLinksCount(accountID int64) (int, error) {

	var count int
	err := repo.DB.QueryRow("select count(*) from links where account_id = $1", accountID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repo *LinksRepository) CreateUserLink(accountID int64, link *Link) (int64, error) {
	var rowID int64
	err := repo.DB.QueryRow(
		"insert into links (short_url, long_url, account_id) values ($1, $2, $3) returning id",
		link.Short, link.Long, accountID,
	).Scan(&rowID)
	return rowID, err
}

func (repo *LinksRepository) DeleteUserLink(accountID int64, link string) (int64, error) {
	var rowID int64
	err := repo.DB.QueryRow(
		"delete from links WHERE short_url = $1 AND account_id = $2 returning id", link, accountID,
	).Scan(&rowID)
	return rowID, err
}

func (repo *LinksRepository) AddUrlToGroup(groupID, linkID int64) error {
	_, err := repo.DB.Exec(`
		insert into links_groups (group_id, link_id) values ($1, $2)
	`, groupID, linkID)
	return err
}

func (repo *LinksRepository) DeleteUrlFromGroup(groupID, linkID int64) error {
	_, err := repo.DB.Exec(`
		delete from links_groups where group_id = $1 and link_id = $2
	`, groupID, linkID)
	return err
}
