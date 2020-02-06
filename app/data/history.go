package data

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"

	"shortly/app/billing"
	"shortly/utils"
)

func incrementTimeSeriesCounter(bucket *bolt.Bucket) error {

	dayRounded := utils.DayNow()

	key := dayRounded.Format(time.RFC3339)

	counter := string(bucket.Get([]byte(key)))
	var intCounter int64
	if counter != "" {
		intCounter, _ = strconv.ParseInt(counter, 0, 64)
	}

	intCounter += 1

	return bucket.Put([]byte(key), []byte(strconv.Itoa(int(intCounter))))
}

type LinkInfo struct {
	Referrers map[string]int
	Locations map[string]int
}

func updateLinkInfo(info LinkRequestData, bucket *bolt.Bucket) error {

	dayRounded := utils.DayNow()
	key := dayRounded.Format(time.RFC3339)
	linkInfo := bucket.Get([]byte(key))

	linkInfoData := &LinkInfo{}
	if len(linkInfo) > 0 {
		if err := json.NewDecoder(bytes.NewBuffer(linkInfo)).Decode(linkInfoData); err != nil {
			return err
		}
	} else {
		linkInfoData.Referrers = make(map[string]int)
		linkInfoData.Locations = make(map[string]int)
	}

	linkInfoData.Locations[info.Location] += 1
	linkInfoData.Referrers[info.Referrer] += 1

	bf := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(bf).Encode(&linkInfoData); err != nil {
		return err
	}

	return bucket.Put([]byte(key), bf.Bytes())
}

// LinkDetail ...
type LinkDetail struct {
	AccountID int64
}

// HistoryDB ...
type HistoryDB struct {
	*bolt.DB
	Limiter *billing.BillingLimiter
	Logger  *log.Logger
}

// LinkDetailsNotFound ...
var LinkDetailsNotFound = errors.New("link details not found")

// DeleteClicks ...
func (d *HistoryDB) DeleteClicks(link string) error {
	return d.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("clicks:" + link))
	})
}

// InsertClick ...
func (d *HistoryDB) InsertClick(link string, t time.Time, counter int) error {
	return d.Update(func(tx *bolt.Tx) error {
		clicksBucket, err := tx.CreateBucketIfNotExists([]byte("clicks:" + link))
		if err != nil {
			return err
		}
		key := t.Format(time.RFC3339)
		return clicksBucket.Put([]byte(key), []byte(strconv.Itoa(counter)))
	})
}

// DeleteInfos ...
func (d *HistoryDB) DeleteInfos(link string) error {
	return d.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte("info:" + link))
	})
}

func (d *HistoryDB) InsertInfo(link string, t time.Time, linkInfo LinkInfo) error {
	return d.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("info:" + link))
		if err != nil {
			return err
		}
		bf := bytes.NewBuffer([]byte{})
		if err := json.NewEncoder(bf).Encode(&linkInfo); err != nil {
			return nil
		}
		key := t.Format(time.RFC3339)
		return bucket.Put([]byte(key), bf.Bytes())
	})
}

type LinkRequestData struct {
	Location string
	Referrer string
}

// Insert ...
func (d *HistoryDB) Insert(link string, info LinkRequestData, r *http.Request) error {

	ipAddr := utils.GetIPAdress(r)

	err := d.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("details"))

		linkB := b.Get([]byte(link))
		if linkB == nil {
			d.Logger.Printf("link(%s) details not found\n", link)
			return nil
		}

		var linkDetail LinkDetail

		if err := json.NewDecoder(bytes.NewBuffer(linkB)).Decode(&linkDetail); err != nil {
			return err
		}

		clicksBucket, err := tx.CreateBucketIfNotExists([]byte("clicks:" + link))
		if err != nil {
			return err
		}

		if err := incrementTimeSeriesCounter(clicksBucket); err != nil {
			return err
		}

		uniqueBucket, err := tx.CreateBucketIfNotExists([]byte("unique:" + ipAddr + ":" + link))
		if err != nil {
			return err
		}

		if err := incrementTimeSeriesCounter(uniqueBucket); err != nil {
			return err
		}

		linkDataBucket, err := tx.CreateBucketIfNotExists([]byte("info:" + link))
		if err != nil {
			return err
		}

		if err := updateLinkInfo(info, linkDataBucket); err != nil {
			return err
		}

		return nil
	})

	return err
}

// InsertDetail ...
func (db *HistoryDB) InsertDetail(shortURL string, accountID int64) error {
	return db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("details"))

		row := LinkDetail{
			AccountID: accountID,
		}

		buffer := bytes.NewBuffer([]byte{})

		if err := json.NewEncoder(buffer).Encode(&row); err != nil {
			return err
		}

		return b.Put([]byte(shortURL), buffer.Bytes())
	})
}

// CounterData ...
type CounterData struct {
	Time  time.Time
	Count int64
}

// LinkInfoData ...
type LinkInfoData struct {
	Time time.Time
	Info LinkInfo
}

// HistoryQueryOption ...
type HistoryQueryOption struct {
	Limit int64
}

type LinkStatistics struct {
	Clicks []CounterData
	Infos  []LinkInfoData
}

// Limit ...
func Limit(limit int64) HistoryQueryOption {
	return HistoryQueryOption{
		Limit: limit,
	}
}

// GetClicksData ...
func (db *HistoryDB) GetClicksData(accountID int64, link string, start, end time.Time, options ...HistoryQueryOption) (*LinkStatistics, error) {

	dataStoreLimit, err := db.Limiter.GetOptionValue("timedata_limit", accountID)
	if err != nil {
		return nil, err
	}

	var dayToStore int64
	for _, opt := range options {
		if opt.Limit > 0 {
			db.Logger.Printf("fetch link(%s) data with override limit: %v\n", link, opt.Limit)
			dayToStore = opt.Limit
		}
	}

	if dayToStore == 0 {
		dayToStore, _ = strconv.ParseInt(dataStoreLimit.Value, 0, 64)
	}

	dayRequested := int64((end.Unix() - start.Unix()) / (3600 * 24))

	if dayToStore < dayRequested {
		return nil, errors.New("limit error")
	}

	var counters []CounterData
	var infos []LinkInfoData

	err = db.View(func(tx *bolt.Tx) error {

		linkBucket := tx.Bucket([]byte("clicks:" + link))

		if linkBucket == nil {
			db.Logger.Printf("history - link(%s) bucket not found\n", link)
			return nil
		}

		b := linkBucket.Cursor()

		if b == nil {
			db.Logger.Printf("history - link(%s) cursor empty\n", link)
			return nil
		}

		startKey := start.Format(time.RFC3339)
		endKey := end.Format(time.RFC3339)

		for k, v := b.Seek([]byte(startKey)); k != nil && bytes.Compare(k, []byte(endKey)) <= 0; k, v = b.Next() {

			timeK, err := time.Parse(time.RFC3339, string(k))
			if err != nil {
				return err
			}

			counterValue, err := strconv.ParseInt(string(v), 0, 64)
			if err != nil {
				return err
			}

			counters = append(counters, CounterData{
				Time:  timeK,
				Count: counterValue,
			})
		}

		db.Logger.Printf("history - fetched interval(%s, %s), found: %v", startKey, endKey, len(counters))

		linkInfoBucket := tx.Bucket([]byte("info:" + link))

		if linkInfoBucket == nil {
			db.Logger.Printf("history - link(%s) info bucket not found\n", link)
			return nil
		}

		linkInfoBucketCursor := linkInfoBucket.Cursor()

		if linkInfoBucketCursor == nil {
			db.Logger.Printf("history - link(%s) info cursor empty\n", link)
			return nil
		}

		for k, v := linkInfoBucketCursor.Seek([]byte(startKey)); k != nil && bytes.Compare(k, []byte(endKey)) <= 0; k, v = linkInfoBucketCursor.Next() {

			timeK, err := time.Parse(time.RFC3339, string(k))
			if err != nil {
				return err
			}

			var linkInfo LinkInfo

			if err := json.NewDecoder(bytes.NewBuffer(v)).Decode(&linkInfo); err != nil {
				return err
			}

			infos = append(infos, LinkInfoData{
				Time: timeK,
				Info: linkInfo,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &LinkStatistics{Clicks: counters, Infos: infos}, nil
}
