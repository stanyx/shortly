package tags

import (
	"database/sql"
	"log"

	"shortly/app/links"
)

type Repository interface {
	GetAllLinkTags(linkID int64) ([]string, error)
	GetAllLinksForTags(tags []string) ([]links.Link, error)
	AddTagToLink(linkID int64, tag *Tag) (int64, error)
	UpdateTagName(linkID int64, oldTagName, newTagName string) error
	DeleteTagFromLink(linkID int64, tagName string) (int64, error)
}

// TagsRepository ...
type TagsRepository struct {
	DB     *sql.DB
	Logger *log.Logger
}

// GetAllLinkTags returns list of all tags for a specified short link
func (r *TagsRepository) GetAllLinkTags(linkID int64) ([]string, error) {
	rows, err := r.DB.Query(
		"select tag from tags where link_id = $1",
		linkID,
	)
	if err != nil {
		return nil, err
	}

	var tags []string

	for rows.Next() {
		var tag string
		err := rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tags, nil
}

// GetAllLinksForTags returns list of Links for a list of tags
func (r *TagsRepository) GetAllLinksForTags(tags []string) ([]links.Link, error) {
	rows, err := r.DB.Query(
		`select distinct(id, short_url, long_url) from links u
		inner join tags t on t.link_id = u.id
		where t.tag IN ($1)
		`, tags)

	if err != nil {
		return nil, err
	}

	var list []links.Link

	for rows.Next() {
		var u links.Link
		err := rows.Scan(&u.ID, &u.Short, &u.Long)
		if err != nil {
			return nil, err
		}
		list = append(list, u)
	}

	defer rows.Close()

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, err
}

// AddTagToLink adds tag to a short link
func (r *TagsRepository) AddTagToLink(linkID int64, tag *Tag) (int64, error) {
	var rowID int64

	tx, err := r.DB.Begin()
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec("select 1 from links where id = $1", linkID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	err = tx.QueryRow(
		"insert into tags (tag, link_id) values ($1, $2) returning id",
		tag.Name, linkID,
	).Scan(&rowID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	return rowID, err
}

// UpdateTagName update tag name for a short link
func (r *TagsRepository) UpdateTagName(linkID int64, oldTagName, newTagName string) error {
	var rowID int64

	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("select 1 from links where id = $1", linkID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.QueryRow(
		"update tags set tag = $1 where tag = $2 and link_id = $3",
		newTagName, oldTagName, linkID,
	).Scan(&rowID)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	return nil
}

// DeleteTagFromLink deletes tag from a short link
func (r *TagsRepository) DeleteTagFromLink(linkID int64, tagName string) (int64, error) {
	var rowID int64
	err := r.DB.QueryRow(
		"delete from tags where link_id = $1 and tag = $2 returning id",
		linkID, tagName,
	).Scan(&rowID)
	return rowID, err
}
