package maintance

import (
	"database/sql"
	"io/ioutil"
	"os"
	"path/filepath"
)

func EnsureGeoIPDatabase(db *sql.DB, databasePath string) error {

	geoIPDBPath := filepath.Join(databasePath, "GeoLite2-Country", "GeoLite2-Country.mmdb")
	_, err := os.Stat(geoIPDBPath)
	if err == nil {
		return nil
	}

	var fileContent []byte

	err = db.QueryRow("select content from files where name='geo2ip' order by downloaded_at desc limit 1").Scan(
		&fileContent,
	)

	if err == sql.ErrNoRows {
		return nil
	} else if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(geoIPDBPath), os.ModePerm); err != nil {
		return err
	}

	return ioutil.WriteFile(geoIPDBPath, fileContent, os.ModePerm)
}
