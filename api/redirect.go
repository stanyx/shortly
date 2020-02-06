package api

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/oschwald/geoip2-golang"

	"shortly/app/data"
	"shortly/cache"
	"shortly/utils"

	"shortly/api/response"

	"shortly/app/links"
)

type LinkRedirect struct {
	ShortUrl string
	LongUrl  string
	Headers  http.Header
	IPAddr   string
	Country  string
	Referer  string
}

// Redirect ...
// @Summary Redirect from short link to associated long url
// @Tags Links
// @ID redirect-short-link
// @Success 307
// @Success 308
// @Failure 400
// @Failure 500
// @Router /* [get]
func Redirect(repo links.ILinksRepository, redirectLogger utils.DbLogger, historyDB *data.HistoryDB, urlCache cache.UrlCache, logger *log.Logger, geoipDbPath string) http.HandlerFunc {

	var geoipDB *geoip2.Reader

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ipAddr := utils.GetIPAdress(r)
		var country string
		var err error

		if geoipDB == nil {
			geoipDB, _ = geoip2.Open(filepath.Join(geoipDbPath, "GeoLite2-Country", "GeoLite2-Country.mmdb"))
		}

		if geoipDB != nil {
			countryObject, err := geoipDB.Country(net.ParseIP(ipAddr))
			if err == nil {
				country = countryObject.Country.Names["en"]
			}
		}

		logger.Printf("redirect start, id_addr = %s, country = %s\n", ipAddr, country)

		shortURL := strings.TrimPrefix(r.URL.Path, "/")

		scheme := r.URL.Scheme
		if scheme == "" {
			scheme = "http://"
		}

		if shortURL == "/" || shortURL == "" {
			http.Redirect(w, r, scheme+r.Host+"/static/index.html", http.StatusPermanentRedirect)
			return
		}

		if shortURL == "admin" {
			http.Redirect(w, r, scheme+r.Host+"/static/admin/index.html", http.StatusPermanentRedirect)
			return
		}

		cacheURLValue, ok := urlCache.Load(shortURL)

		var longURL string

		if ok {
			longURL, ok = cacheURLValue.(string)
			if !ok {
				response.Text(w, "url is not a string", http.StatusBadRequest)
				return
			}
		} else {
			longURL, err = repo.UnshortenURL(shortURL)
			if err == nil {
				logger.Printf("cache miss, short=%v, long=%v\n", shortURL, longURL)
				urlCache.Store(shortURL, longURL)
			}
		}

		if longURL == "" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("not found"))
			return
		}

		if !(strings.HasPrefix(longURL, "http") || strings.HasPrefix(longURL, "https")) {
			longURL = "https://" + longURL
		}

		validURL, err := url.Parse(longURL)
		if err != nil {
			response.Text(w, "url has incorrect format", http.StatusBadRequest)
			return
		}

		referrers := r.Header[http.CanonicalHeaderKey("Referer")]

		var referer string
		if len(referrers) > 0 {
			referer = referrers[0]
		}

		requestData := data.LinkRequestData{
			Location: country,
			Referrer: referer,
		}

		if err := historyDB.Insert(shortURL, requestData, r); err != nil {
			logError(logger, err)
			response.Text(w, "internal server error", http.StatusInternalServerError)
			return
		}

		body, err := json.Marshal(&LinkRedirect{
			ShortUrl: shortURL,
			LongUrl:  validURL.String(),
			Headers:  r.Header,
			IPAddr:   ipAddr,
			Country:  country,
			Referer:  referer,
		})
		if err != nil {
			logError(logger, err)
			response.Text(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if err := redirectLogger.Push(body); err != nil {
			logError(logger, err)
			response.Text(w, "internal server error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, validURL.String(), http.StatusSeeOther)

	})
}

type IPInfo struct {
	Country string
}

func GetIPInfo(geoipDbPath string) http.HandlerFunc {

	var geoipDB *geoip2.Reader

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ipAddr := utils.GetIPAdress(r)
		var country string

		if geoipDB == nil {
			geoipDB, _ = geoip2.Open(filepath.Join(geoipDbPath, "GeoLite2-Country", "GeoLite2-Country.mmdb"))
		}

		if geoipDB != nil {
			countryObject, err := geoipDB.Country(net.ParseIP(ipAddr))
			if err == nil {
				country = countryObject.Country.Names["en"]
			}
		}

		response.Object(w, &IPInfo{Country: country}, http.StatusOK)

	})
}
