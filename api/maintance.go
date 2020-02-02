package api

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"shortly/api/response"
	"shortly/app/billing"
	"shortly/utils"

	"github.com/stripe/stripe-go"
)

// UpdateGeoIPDatabase ...
func UpdateGeoIPDatabase(downloadURL, geoIPDatabasePath, key string, logger *log.Logger) http.HandlerFunc {

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

		if _, err := utils.Unzip(tf.Name(), geoIPDatabasePath); err != nil {
			logError(logger, err)
			response.Error(w, "unzip file error", http.StatusInternalServerError)
			return
		}

		response.Ok(w)
	})

}

// UploadGeoIPDatabase ...
func UploadGeoIPDatabase(uploadPath string, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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

		if _, err := utils.Unzip(tf.Name(), uploadPath); err != nil {
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
