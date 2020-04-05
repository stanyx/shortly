package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// AuditRequest ...
type AuditRequest struct {
	//TODO
}

// AuditQuery ...
type AuditQuery struct {
}

func (i *AuditQuery) doInsertQuery(entityName string, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	var rowID int64
	err := tx.QueryRow(query, args...).Scan(&rowID)
	if err != nil {
		return 0, errors.Wrap(err, "insert error")
	}
	_, err = tx.Exec(fmt.Sprintf(`
		insert into audit (entity, entity_id, snapshot, timestamp, action) values ($1, $2, 
			(select row_to_json(e, true) from %s as e where id = $2), now(), 'create'
		)
	`, entityName), entityName, rowID)
	return rowID, errors.Wrap(err, "audit error")
}

func (i *AuditQuery) doUpdateQuery(entityID int64, entityName string, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	var rowID int64
	var oldSnapshotStr, newSnapshotStr string

	err := tx.QueryRow(
		fmt.Sprintf("select row_to_json(e, true) from %s as e where id = $1",
			entityName),
		entityID,
	).Scan(&oldSnapshotStr)

	oldSnapshot := make(map[string]interface{})
	err = json.Unmarshal([]byte(oldSnapshotStr), &oldSnapshot)
	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(query, args...).Scan(&rowID)
	if err != nil {
		return 0, errors.Wrap(err, "insert error")
	}

	err = tx.QueryRow(
		fmt.Sprintf("select row_to_json(e, true) from %s as e where id = $1",
			entityName),
		entityID,
	).Scan(&newSnapshotStr)

	newSnapshot := make(map[string]interface{})
	err = json.Unmarshal([]byte(newSnapshotStr), &newSnapshot)
	if err != nil {
		return 0, err
	}

	diff := make(map[string][]interface{})

	for k, v := range oldSnapshot {
		if _, ok := diff[k]; !ok {
			diff[k] = []interface{}{v, nil}
		}
	}

	for k, v := range newSnapshot {
		if _, ok := diff[k]; !ok {
			diff[k] = []interface{}{nil, v}
		} else {
			diff[k][1] = v
		}
	}

	diffBytes, err := json.Marshal(diff)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(fmt.Sprintf(`
		insert into audit (entity, entity_id, diff, snapshot, timestamp, action) values ($1, $2, $3,
			(select row_to_json(e, true) from %s as e where id = $2), now(), 'update'
		)
	`, entityName), entityName, rowID, string(diffBytes))
	return rowID, errors.Wrap(err, "audit error")
}

func (i *AuditQuery) doDeleteQuery(entityName string, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	var rowID int64
	var snapshot string

	snapshotQuery := strings.Replace(query, "delete", "select row_to_json(e)", -1)
	snapshotQuery = strings.Replace(snapshotQuery, "returning id", "", -1)
	fmt.Println(snapshotQuery)
	if err := tx.QueryRow(snapshotQuery, args...).Scan(&snapshot); err != nil {
		return 0, err
	}
	err := tx.QueryRow(query, args...).Scan(&rowID)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(`
		insert into audit (entity, entity_id, snapshot, timestamp, action) values ($1, $2, $3,
			now(), 'delete'
		)
	`, entityName, rowID, snapshot)
	return rowID, err
}

func (i *AuditQuery) Create(entityName string, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	return i.doInsertQuery(entityName, tx, query, args...)
}

func (i *AuditQuery) Delete(entityName string, tx *sql.Tx, query string, args ...interface{}) (int64, error) {
	return i.doDeleteQuery(entityName, tx, query, args...)
}

func (i *AuditQuery) CreateTx(entityName string, db *sql.DB, query string, args ...interface{}) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	rowID, err := i.doInsertQuery(entityName, tx, query, args...)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return rowID, nil
}

func (i *AuditQuery) UpdateTx(entityID int64, entityName string, db *sql.DB, query string, args ...interface{}) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	rowID, err := i.doUpdateQuery(entityID, entityName, tx, query, args...)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return rowID, nil
}

func (i *AuditQuery) DeleteTx(entityName string, db *sql.DB, query string, args ...interface{}) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	rowID, err := i.doDeleteQuery(entityName, tx, query, args...)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return rowID, nil
}
