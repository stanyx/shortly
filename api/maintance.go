package api

import (
	"bufio"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"shortly/api/response"
	"shortly/app/billing"
	"shortly/utils"

	"github.com/stripe/stripe-go"
)

// UpdateGeoIPDatabase ...
func UpdateGeoIPDatabase(db *sql.DB, downloadURL, geoIPDatabasePath, key string, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		oldDatabases, err := filepath.Glob(filepath.Join(geoIPDatabasePath, "GeoLite2-*"))
		if err != nil {
			logError(logger, err)
			fmt.Fprintf(w, "internal server error")
			return
		}

		for _, oldDb := range oldDatabases {
			if err := os.RemoveAll(oldDb); err != nil {
				logError(logger, err)
				fmt.Fprintf(w, "internal server error")
				return
			}
		}

		urlWithToken := fmt.Sprintf(downloadURL, key)
		fmt.Printf("download geo ip database, url=%v, path=%v\n", urlWithToken, geoIPDatabasePath)

		client := &http.Client{Timeout: time.Second * 10}

		tf, err := ioutil.TempFile(filepath.Base(geoIPDatabasePath), "geoip_db")
		if err != nil {
			logError(logger, err)
			fmt.Fprintf(w, "internal server error")
			return
		}

		defer func() {
			if err := os.Remove(tf.Name()); err != nil {
				logError(logger, err)
				return
			}
		}()
		defer tf.Close()

		rq, err := http.NewRequest("GET", urlWithToken, nil)
		if err != nil {
			logError(logger, err)
			response.Error(w, "request error", http.StatusInternalServerError)
			return
		}

		resp, err := client.Do(rq)
		if err != nil {
			logError(logger, err)
			response.Error(w, "download error", http.StatusInternalServerError)
			return
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return
			}
			logger.Println("error response", string(body))
			response.Error(w, "download error, status_code != 200", http.StatusInternalServerError)
			return
		}

		if _, err := io.Copy(tf, resp.Body); err != nil {
			logError(logger, err)
			response.Error(w, "body copy error", http.StatusInternalServerError)
			return
		}

		_, err = utils.Untar(tf.Name(), geoIPDatabasePath)
		if err != nil {
			logError(logger, err)
			response.Error(w, "unzip file error", http.StatusInternalServerError)
			return
		}

		matches, err := filepath.Glob(filepath.Join(filepath.Dir(tf.Name()), "GeoLite2-Country_*"))
		if err != nil {
			logError(logger, err)
			response.Error(w, "rename file error", http.StatusInternalServerError)
			return
		}

		re, err := regexp.Compile(`_\d+`)
		if err != nil {
			logError(logger, err)
			response.Error(w, "rename file error", http.StatusInternalServerError)
			return
		}

		var geo2ipDatabaseName string

		for i, m := range matches {
			geo2ipDatabaseName = re.ReplaceAllString(m, "")
			if err := os.Rename(m, geo2ipDatabaseName); err != nil {
				logError(logger, err)
				response.Error(w, "rename file error", http.StatusInternalServerError)
				return
			}

			f, err := os.Open(filepath.Join(geo2ipDatabaseName, "GeoLite2-Country.mmdb"))
			if err != nil {
				logError(logger, err)
				response.Error(w, "open file error", http.StatusInternalServerError)
				return
			}

			fileContent, err := ioutil.ReadAll(f)
			if err != nil {
				f.Close()
				logError(logger, err)
				response.Error(w, "read file error", http.StatusInternalServerError)
				return
			}

			f.Close()

			_, err = db.Exec("insert into files (name, content, downloaded_at) values ($1, $2, now())", "geo2ip", fileContent)
			if err != nil {
				logError(logger, err)
				response.Error(w, "insert data error", http.StatusInternalServerError)
			}

			if i > 0 {
				break
			}
		}

		response.Ok(w)
	})

}

// UploadGeoIPDatabase ...
func UploadGeoIPDatabase(uploadPath string, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		dbType := "tar"
		dbTypeArgs := r.URL.Query()["dbType"]
		if len(dbTypeArgs) > 0 {
			dbType = dbTypeArgs[0]
		}

		if !(dbType == "csv" || dbType == "tar") {
			fmt.Fprintf(w, "file type is invalid")
			return
		}

		if err := r.ParseMultipartForm(10 << 20); err != nil {
			logError(logger, err)
			fmt.Fprintf(w, "parse form error")
			return
		}

		file, _, err := r.FormFile("file")
		if err != nil {
			logError(logger, err)
			fmt.Fprintf(w, "internal server error")
			return
		}
		defer file.Close()

		oldDatabases, err := filepath.Glob(filepath.Join(uploadPath, "GeoLite2-*"))
		if err != nil {
			logError(logger, err)
			fmt.Fprintf(w, "internal server error")
			return
		}

		for _, oldDb := range oldDatabases {
			if err := os.RemoveAll(oldDb); err != nil {
				logError(logger, err)
				fmt.Fprintf(w, "internal server error")
				return
			}
		}

		tf, err := ioutil.TempFile(filepath.Base(uploadPath), "geoip_db")
		if err != nil {
			logError(logger, err)
			fmt.Fprintf(w, "internal server error")
			return
		}

		if _, err := io.Copy(tf, file); err != nil {
			logError(logger, err)
			fmt.Fprintf(w, "internal server error")
			return
		}

		if dbType == "csv" {
			if _, err := utils.Unzip(tf.Name(), uploadPath); err != nil {
				logError(logger, err)
				fmt.Fprintf(w, "internal server error")
				return
			}
		} else if dbType == "tar" {
			if _, err := utils.Untar(tf.Name(), uploadPath); err != nil {
				logError(logger, err)
				fmt.Fprintf(w, "internal server error")
				return
			}
		}

		defer func() {
			if err := os.Remove(tf.Name()); err != nil {
				logError(logger, err)
				return
			}
		}()
		defer tf.Close()

		fmt.Fprintf(w, "success")

	})
}

// LoadStripeFixtures ...
func LoadStripeFixtures(repo *billing.BillingRepository, logger *log.Logger) http.HandlerFunc {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		plansCsvFile, err := os.Open("./fixtures/stripe_fixtures_plans.csv")
		if err != nil {
			logError(logger, err)
			response.Error(w, "open plans csv error", http.StatusInternalServerError)
			return
		}

		defer plansCsvFile.Close()

		reader := csv.NewReader(bufio.NewReader(plansCsvFile))

		for {
			line, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				logError(logger, err)
				response.Error(w, "read plans csv error", http.StatusInternalServerError)
				return
			}
			createdTime, _ := strconv.ParseInt(line[3], 0, 64)
			if err := repo.UpsertStripePlan(&stripe.Plan{ID: line[1], Nickname: line[2], Created: createdTime}); err != nil {
				logError(logger, err)
				response.Error(w, "upsert plans error", http.StatusInternalServerError)
				return
			}
		}

		productsCsvFile, err := os.Open("./fixtures/stripe_fixtures_plans.csv")
		if err != nil {
			logError(logger, err)
			response.Error(w, "open plans csv error", http.StatusInternalServerError)
			return
		}

		defer productsCsvFile.Close()

		reader = csv.NewReader(bufio.NewReader(productsCsvFile))

		for {
			line, err := reader.Read()
			if err == io.EOF {
				break
			} else if err != nil {
				logError(logger, err)
				response.Error(w, "read plans csv error", http.StatusInternalServerError)
				return
			}
			createdTime, _ := strconv.ParseInt(line[3], 0, 64)
			if err := repo.UpsertStripeProduct(&stripe.Product{ID: line[1], Name: line[2], Created: createdTime}); err != nil {
				logError(logger, err)
				response.Error(w, "upsert plans error", http.StatusInternalServerError)
				return
			}
		}

		response.Ok(w)
	})
}
