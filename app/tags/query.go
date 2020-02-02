package tags

import (
	"database/sql"
	"log"

	"shortly/app/links"
)

// TagsRepository ...
type TagsRepository struct {
	DB     *sql.DB
	Logger *log.Logger
}

// GetAllLinkTags ...
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

// GetAllLinksForTags ...
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

// AddTagToLink ...
func (r *TagsRepository) AddTagToLink(linkID int64, tag string) (int64, error) {
	var rowID int64
	err := r.DB.QueryRow(
		"insert into tags (tag, link_id) values ($1, $2) returning id",
		tag, linkID,
	).Scan(&rowID)
	return rowID, err
}

// DeleteTagFromLink ...
func (r *TagsRepository) DeleteTagFromLink(linkID int64, tagName string) (int64, error) {
	var rowID int64
	err := r.DB.QueryRow(
		"delete from tags where link_id = $1 and tag = $2 returning id", linkID, tagName,
	).Scan(&rowID)
	return rowID, err
}
