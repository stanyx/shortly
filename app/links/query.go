package links

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"

	"shortly/utils"
)

type ILinksRepository interface {
	UnshortenURL(string) (string, error)
	GetLinkByID(int64) (Link, error)
	UpdateUserLink(int64, int64, *Link) (*sql.Tx, error)
	GetAllLinks() ([]Link, error)
	GenerateLink() string
	CreateLink(*Link) error
	CreateUserLink(accountID int64, link *Link) (*sql.Tx, int64, error)
	DeleteUserLink(accountID int64, linkID int64) (*sql.Tx, int64, error)
	GetUserLinks(accountID, userID int64, filters ...LinkFilter) ([]Link, error)
	GetUserLinksCount(accountID int64, startTime, endTime time.Time) (int, error)
	AddUrlToGroup(groupID int64, linkID int64) error
	DeleteUrlFromGroup(groupID int64, linkID int64) error
}

type LinksRepository struct {
	DB        *sql.DB
	Logger    *log.Logger
	callbacks map[string]func(int64, interface{})
}

func (repo *LinksRepository) addCallback(name string, f func(int64, interface{})) {
	if repo.callbacks == nil {
		repo.callbacks = make(map[string]func(int64, interface{}))
	}
	repo.callbacks[name] = f
}

func (repo *LinksRepository) callback(name string, accountID int64, payload interface{}) {
	cb, ok := repo.callbacks[name]
	if !ok {
		return
	}
	cb(accountID, payload)
}

func (repo *LinksRepository) OnCreate(f func(int64, interface{})) {
	repo.addCallback("Create", f)
}

func (repo *LinksRepository) OnDelete(f func(int64, interface{})) {
	repo.addCallback("Delete", f)
}

func (repo *LinksRepository) OnHide(f func(int64, interface{})) {
	repo.addCallback("Hide", f)
}

func (repo *LinksRepository) UnshortenURL(shortURL string) (string, error) {

	query := "select long_url from links where short_url = $1"

	var longURL string
	err := repo.DB.QueryRow(query, shortURL).Scan(&longURL)
	if err != nil {
		return "", err
	}

	return longURL, nil
}

func (repo *LinksRepository) GetAllLinks() ([]Link, error) {

	query := "select short_url, long_url, account_id from links"
	var queryArgs []interface{}
	rows, err := repo.DB.Query(query, queryArgs...)
	if err != nil {
		return nil, err
	}

	var list []Link

	for rows.Next() {
		var shortURL, longURL string
		var accountID int64
		err := rows.Scan(&shortURL, &longURL, &accountID)
		if err != nil {
			return nil, err
		}
		list = append(list, Link{
			AccountID: accountID,
			Short:     shortURL,
			Long:      longURL,
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
	ShortUrl []string
	LongUrl  []string
	Tags     []string
	FullText string
	LinkID   int64
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
	select u.id, u.short_url, u.long_url, u.description, u.tl, u.hide
	from (
		select *, t.tags_list tl from links
		left join url_group ug on ug.link_id = links.id
		left outer join url_tags t on t.link_id = links.id
		where (links.account_id = $1 and not exists (select 1 from url_group)) 
		or (ug.link_id is not null and exists (select 1 from url_group))
	) u
	order by u.id desc
	`

	queryArgs := []interface{}{accountID, userID}

	var filterExpressions []string

	for _, f := range filters {
		if len(f.Tags) > 0 {
			exp := []string{fmt.Sprintf("u.tl && $%d", len(queryArgs)+1)}
			queryArgs = append(queryArgs, pq.Array(f.Tags))
			filterExpressions = append(filterExpressions, fmt.Sprintf("(%s)", strings.Join(exp, " OR ")))
		}
		if len(f.ShortUrl) > 0 {
			exp := []string{}
			for _, v := range f.ShortUrl {
				exp = append(exp, fmt.Sprintf("u.short_url LIKE $%d", len(queryArgs)+1))
				queryArgs = append(queryArgs, v+"%")
			}
			filterExpressions = append(filterExpressions, fmt.Sprintf("(%s)", strings.Join(exp, " OR ")))
		}
		if len(f.LongUrl) > 0 {
			exp := []string{}
			for _, v := range f.LongUrl {
				exp = append(exp, fmt.Sprintf("u.long_url LIKE $%d", len(queryArgs)+1))
				queryArgs = append(queryArgs, v+"%")
			}
			filterExpressions = append(filterExpressions, fmt.Sprintf("(%s)", strings.Join(exp, " OR ")))
		}
		if f.FullText != "" {
			exp := []string{
				fmt.Sprintf("u.tl && $%d", len(queryArgs)+1),
				fmt.Sprintf("u.short_url LIKE $%d", len(queryArgs)+2),
				fmt.Sprintf("u.long_url LIKE $%d", len(queryArgs)+3),
			}
			queryArgs = append(queryArgs, pq.Array([]string{f.FullText}))
			queryArgs = append(queryArgs, f.FullText+"%")
			queryArgs = append(queryArgs, f.FullText+"%")
			filterExpressions = append(filterExpressions, fmt.Sprintf("(%s)", strings.Join(exp, " OR ")))
		}
		if f.LinkID > 0 {
			filterExpressions = append(filterExpressions, fmt.Sprintf("u.id = $%d", len(queryArgs)+1))
			queryArgs = append(queryArgs, f.LinkID)
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
		err := rows.Scan(&link.ID, &link.Short, &link.Long, &link.Description, pq.Array(&link.Tags), &link.Hidden)
		if err != nil {
			return nil, err
		}
		list = append(list, link)
	}

	return list, nil
}

func (repo *LinksRepository) GetLinkByID(linkID int64) (Link, error) {

	var link Link
	err := repo.DB.QueryRow(`
		 select short_url, long_url, description from "links" where id = $1
	`, linkID).Scan(&link.Short, &link.Long, &link.Description)

	return link, err
}

func (repo *LinksRepository) GetUserLinksCount(accountID int64, createdStartTime, createdEndTime time.Time) (int, error) {

	var count int
	err := repo.DB.QueryRow(`
		select count(*) from links where account_id = $1 and created_at >= $2 and created_at <= $3
	`, accountID, createdStartTime, createdEndTime).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repo *LinksRepository) GenerateLink() string {
	return utils.RandomString(5)
}

func (repo *LinksRepository) CreateUserLink(accountID int64, link *Link) (*sql.Tx, int64, error) {
	var rowID int64
	tx, err := repo.DB.Begin()
	if err != nil {
		return nil, 0, err
	}
	err = tx.QueryRow(
		"insert into links (short_url, long_url, account_id, created_at) values ($1, $2, $3, now()) returning id",
		link.Short, link.Long, accountID,
	).Scan(&rowID)
	if err != nil {
		return nil, 0, err
	}

	repo.callback("Create", accountID, link)
	return tx, rowID, err
}

func (repo *LinksRepository) UpdateUserLink(accountID, linkID int64, link *Link) (*sql.Tx, error) {
	tx, err := repo.DB.Begin()
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(
		"update links set long_url = $1, description = $2 where id = $3 and account_id = $4",
		link.Long, link.Description, linkID, accountID,
	)
	return tx, err
}

func (repo *LinksRepository) DeleteUserLink(accountID int64, linkID int64) (*sql.Tx, int64, error) {
	var rowID int64
	tx, err := repo.DB.Begin()
	if err != nil {
		return nil, 0, err
	}
	err = tx.QueryRow(
		"delete from links WHERE id = $1 AND account_id = $2 returning id", linkID, accountID,
	).Scan(&rowID)
	if err != nil {
		return nil, 0, err
	}
	repo.callback("Delete", accountID, linkID)
	return tx, rowID, err
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

func (repo *LinksRepository) BulkCreateLinks(accountID int64, links []string) ([]Link, error) {

	query := "insert into links (short_url, long_url, account_id) values "
	var queryArgs []interface{}

	var createdLinks []Link
	for i, l := range links {
		if i > 0 {
			query += ", "
		}
		query += fmt.Sprintf("($%v, $%v, $%v)", i*3+1, i*3+2, i*3+3)

		shortURL := repo.GenerateLink()
		queryArgs = append(queryArgs, []interface{}{
			shortURL, l, accountID,
		}...)
		createdLinks = append(createdLinks, Link{
			Short: shortURL,
			Long:  l,
		})
	}

	_, err := repo.DB.Exec(query, queryArgs...)
	return createdLinks, err
}

func (repo *LinksRepository) HideUserLink(accountID int64, linkID int64) (*sql.Tx, error) {

	tx, err := repo.DB.Begin()
	if err != nil {
		return nil, err
	}

	query := "update links set hide = true where id = $1 and account_id = $2"
	queryArgs := []interface{}{linkID, accountID}

	_, err = tx.Exec(query, queryArgs...)
	if err != nil {
		return nil, err
	}

	repo.callback("Hide", accountID, linkID)
	return tx, err
}
