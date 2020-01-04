package billing

import (
	"fmt"
	"log"
	"bytes"
	"errors"
	"encoding/json"
	"strconv"

	bolt "go.etcd.io/bbolt"

	"shortly/app/links"
)

var LimitExceededError = errors.New("limit exceeded error")

type BillingLimiter struct {
	UrlRepo *links.LinksRepository
	Repo    *BillingRepository
	DB      *bolt.DB
	Logger  *log.Logger
}

func (l *BillingLimiter) SetPlanOptions(accountID int64, options []BillingOption) error {

	return l.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("billing"))

		buff := bytes.NewBuffer([]byte{})
		if err := json.NewEncoder(buff).Encode(&options); err != nil {
			return err
		}

		err := b.Put([]byte(fmt.Sprintf("%v", accountID)), buff.Bytes())
		return err
	})

}

func (l *BillingLimiter) LoadData() error {

	plans, err := l.Repo.GetAllUserBillingPlans()
	if err != nil {
		return err
	}

	for _, p := range plans {

		for i, opt := range p.Options {
			switch opt.Name {
			case "url_limit":

				cnt, err := l.UrlRepo.GetUserLinksCount(p.AccountID)
				if err != nil {
					return err
				}
				topCount, _ := strconv.ParseInt(p.Options[i].Value, 0, 64)
				l.Logger.Printf("(user=%v) set billing value(%s) to %v, max=%v, value=%v", p.AccountID, opt.Name, topCount - int64(cnt), topCount, cnt)
				p.Options[i].Value = fmt.Sprintf("%v", topCount - int64(cnt))
			case "timedata_limit":
			default:
				return errors.New("billing option is not supported")
			}
		}

		if err := l.SetPlanOptions(p.AccountID, p.Options); err != nil {
			return err
		}
	}

	return nil
}

var OptionNotFound = errors.New("option not found")

func (l *BillingLimiter) GetOptionValue(optionName string, accountID int64) (*BillingOption, error) {
	var optionValue *BillingOption
	err := l.DB.View(func(tx *bolt.Tx) error {
		v, err := l.GetValue(tx, optionName, accountID)
		if err == nil {
			optionValue = v
		}
		return err
	})
	return optionValue, err
}

func (l *BillingLimiter) GetValue(tx *bolt.Tx, optionName string, accountID int64) (*BillingOption, error) {

	b := tx.Bucket([]byte("billing"))
	v := b.Get([]byte(fmt.Sprintf("%v", accountID)))

	if len(v) == 0 {
		return nil, OptionNotFound
	}

	buff := bytes.NewBuffer(v)

	var options []BillingOption
	if err := json.NewDecoder(buff).Decode(&options); err != nil {
		return nil, err
	}

	for _, option := range options {
		if option.Name == optionName {
			return &option, nil
		}
	}

	return nil, OptionNotFound
}

func (l *BillingLimiter) CheckLimits(optionName string, accountID int64) error {

	return l.DB.Update(func(tx *bolt.Tx) error {

		targetOption, err := l.GetValue(tx, optionName, accountID)
		if err == OptionNotFound {
			return LimitExceededError
		} else if err != nil {
			return err
		}

		value, _ := strconv.ParseInt(targetOption.Value, 0, 64)
		l.Logger.Printf("(user=%v) current billing value(%s) value: %v\n", accountID, optionName, value)
		if value <= 0 {
			return LimitExceededError
		}

		return nil
	})

}

func (l *BillingLimiter) Reduce(optionName string, accountID int64) error {

	return l.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte("billing"))
		v := b.Get([]byte(fmt.Sprintf("%v", accountID)))

		if len(v) == 0 {
			return LimitExceededError
		}

		buff := bytes.NewBuffer(v)

		var options []BillingOption
		if err := json.NewDecoder(buff).Decode(&options); err != nil {
			return err
		}

		var targetOption *BillingOption
		var optionIndex int

		for i, option := range options {
			if option.Name == optionName {
				targetOption = &option
				optionIndex = i
				break
			}
		}

		if targetOption != nil {

			value, _ := strconv.ParseInt(targetOption.Value, 0, 64)

			if value > 0 {

				options[optionIndex].Value = fmt.Sprintf("%v", value - 1)
				buff := bytes.NewBuffer([]byte{})
				if err := json.NewEncoder(buff).Encode(&options); err != nil {
					return err
				}

				err := b.Put([]byte(fmt.Sprintf("%v", accountID)), buff.Bytes())
				if err != nil {
					return err
				}

				l.Logger.Printf("(user=%v) reduce billing value(%s) to %v", accountID, optionName, value - 1)
				return nil
			}

		} else {
			return LimitExceededError
		}

		return nil
	})

}