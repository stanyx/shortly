package billing

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"

	"shortly/app/links"
)

const billingDatabaseName = "billing"

var LimitExceededError = errors.New("limit exceeded error")

type BillingLimiter struct {
	UrlRepo *links.LinksRepository
	Repo    *BillingRepository
	DB      *bolt.DB
	Logger  *log.Logger

	locks sync.Map
}

func (l *BillingLimiter) Lock(accountID int64) *sync.Mutex {
	var lock sync.Mutex
	v, _ := l.locks.LoadOrStore(accountID, &lock)
	m := v.(*sync.Mutex)
	m.Lock()
	return m
}

func (l *BillingLimiter) resetOptionForInterval(tx *bolt.Tx, option BillingOption, accountID int64, planStartTime, planEndTime time.Time) error {

	switch option.Name {
	case "url_limit":
		cnt, err := l.UrlRepo.GetUserLinksCount(accountID, planStartTime, planEndTime)
		if err != nil {
			return err
		}
		topCount := option.AsInt64()
		l.Logger.Printf("(user=%v) reset billing value(%s) to %v, max=%v, value=%v", accountID, option.Name, topCount-int64(cnt), topCount, cnt)
		if err := l.UpdateOption(tx, option.Name, accountID, func(_ int64) int64 { return topCount - int64(cnt) }); err != nil {
			return err
		}
		return nil
	default:
		return OptionNotFound
	}
}

func (l *BillingLimiter) resetOption(tx *bolt.Tx, optionName string, accountID int64) (*BillingOption, error) {

	plans, err := l.Repo.GetAllUserBillingPlans(accountID)
	if err != nil {
		return nil, err
	}

	if len(plans) != 1 {
		return nil, errors.New("multiple plans found")
	}

	plan := plans[0]

	var option *BillingOption
	for _, opt := range plan.Options {
		if opt.Name == optionName {
			option = &opt
			break
		}
	}

	if option == nil {
		return nil, errors.New("option not found")
	}

	if err := l.resetOptionForInterval(tx, *option, accountID, plan.Start, plan.End); err != nil {
		return nil, err
	}

	return option, nil
}

func (l *BillingLimiter) Reset(optionName string, accountID int64) error {
	return l.DB.Update(func(tx *bolt.Tx) error {
		_, err := l.resetOption(tx, optionName, accountID)
		return err
	})
}

func (l *BillingLimiter) SetPlanOptions(accountID int64, options []BillingOption) error {

	return l.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(billingDatabaseName))

		buff := bytes.NewBuffer([]byte{})
		if err := json.NewEncoder(buff).Encode(&options); err != nil {
			return err
		}

		err := b.Put([]byte(fmt.Sprintf("%v", accountID)), buff.Bytes())
		return err
	})

}

// LoadData fill billing database with actual billing information
func (l *BillingLimiter) LoadData() error {

	plans, err := l.Repo.GetAllUserBillingPlans(0)
	if err != nil {
		return err
	}

	for _, p := range plans {

		for i, opt := range p.Options {
			switch opt.Name {
			case "url_limit":
				cnt, err := l.UrlRepo.GetUserLinksCount(p.AccountID, p.Start, p.End)
				if err != nil {
					return err
				}
				topCount := p.Options[i].AsInt64()
				l.Logger.Printf("(user=%v) set billing value(%s) to %v, max=%v, value=%v", p.AccountID, opt.Name, topCount-int64(cnt), topCount, cnt)
				p.Options[i].Value = fmt.Sprintf("%v", topCount-int64(cnt))
			case "timedata_limit":
			case "users_limit":
			case "tags_limit":
			case "groups_limit":
			case "rate_limit":
			default:
				return fmt.Errorf("billing option is not supported: %v", opt.Name)
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

// GetValue returns an actual billing value for provided option
func (l *BillingLimiter) GetValue(tx *bolt.Tx, optionName string, accountID int64) (*BillingOption, error) {

	b := tx.Bucket([]byte(billingDatabaseName))

	v := b.Get([]byte(fmt.Sprintf("%v", accountID)))
	// cache miss only possible with plan reset
	if len(v) == 0 {
		option, err := l.resetOption(tx, optionName, accountID)
		if err != nil {
			return nil, err
		}
		return option, nil
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

		option, err := l.GetValue(tx, optionName, accountID)
		if err == OptionNotFound {
			return LimitExceededError
		} else if err != nil {
			return err
		}

		value := option.AsInt64()
		l.Logger.Printf("(user=%v) current billing value(%s) value: %v\n", accountID, optionName, value)
		if value <= 0 {
			return LimitExceededError
		}

		return nil
	})

}

func (l *BillingLimiter) UpdateOption(tx *bolt.Tx, optionName string, accountID int64, f func(int64) int64) error {
	b := tx.Bucket([]byte(billingDatabaseName))
	v := b.Get([]byte(fmt.Sprintf("%v", accountID)))

	if len(v) == 0 {
		return LimitExceededError
	}

	buff := bytes.NewBuffer(v)

	var options []BillingOption
	if err := json.NewDecoder(buff).Decode(&options); err != nil {
		return err
	}

	var update bool
	for i, option := range options {
		if option.Name == optionName {
			value, _ := strconv.ParseInt(option.Value, 0, 64)
			updatedValue := f(value)
			if updatedValue < 0 {
				return errors.New("value is below zero")
			}
			l.Logger.Printf("(user=%v) changed billing value(%s) to %v", accountID, optionName, updatedValue)
			options[i].Value = fmt.Sprintf("%v", updatedValue)
			update = true
			break
		}
	}

	if update {
		buff := bytes.NewBuffer([]byte{})
		if err := json.NewEncoder(buff).Encode(&options); err != nil {
			return err
		}
		return b.Put([]byte(fmt.Sprintf("%v", accountID)), buff.Bytes())
	}

	return nil
}

func (l *BillingLimiter) Reduce(optionName string, accountID int64) error {
	return l.DB.Update(func(tx *bolt.Tx) error {
		return l.UpdateOption(tx, optionName, accountID, func(v int64) int64 {
			return v - 1
		})
	})
}

func (l *BillingLimiter) Increase(optionName string, accountID int64) error {
	return l.DB.Update(func(tx *bolt.Tx) error {
		return l.UpdateOption(tx, optionName, accountID, func(v int64) int64 {
			return v + 1
		})
	})
}
