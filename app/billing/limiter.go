package billing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"

	"shortly/app/links"
	"shortly/utils"
)

const billingDatabaseName = "billing"

var (
	// LimitExceededError ...
	LimitExceededError = errors.New("limit exceeded error")
	// BillingAccountExpiredError ...
	BillingAccountExpiredError = errors.New("billing account expired")

	// counters

	Options = []string{
		"url_limit",
		"timedata_limit",
		"users_limit",
		"tags",
		"tags_limit",
		"groups",
		"groups_limit",
		"campaigns",
		"campaigns_limit",
		"rate_limit",
	}
)

// BillingAccount ...
type BillingAccount struct {
	Start   time.Time
	End     time.Time
	Options []BillingOption
}

// BillingLimiter ...
type BillingLimiter struct {
	UrlRepo *links.LinksRepository
	Repo    *BillingRepository
	DB      *bolt.DB
	Logger  *log.Logger

	locks              sync.Map
	defaultAccountPlan AccountBillingPlan
}

// Lock ...
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

// resetOption ...
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
		return nil, fmt.Errorf("option (%s) not found", optionName)
	}

	if err := l.resetOptionForInterval(tx, *option, accountID, plan.Start, plan.End); err != nil {
		return nil, err
	}

	return option, nil
}

// GetBillingStatistics ...
func (l *BillingLimiter) GetBillingStatistics(accountID int64, startTime, endTime time.Time) ([]BillingParameter, error) {

	var stat []BillingParameter

	err := l.DB.View(func(tx *bolt.Tx) error {
		for _, optionName := range Options {
			v, err := l.getOption(tx, optionName, accountID)
			if err == OptionNotFound {
				continue
			} else if err != nil {
				return err
			}

			currentValue := "-"

			switch optionName {
			case "url_limit":
				cnt, err := l.UrlRepo.GetUserLinksCount(accountID, startTime, endTime)
				if err != nil {
					return err
				}
				currentValue = fmt.Sprintf("%v", cnt)
			}

			stat = append(stat, BillingParameter{
				BillingOption: *v,
				CurrentValue:  currentValue,
			})

		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return stat, nil
}

// Reset ...
func (l *BillingLimiter) Reset(optionName string, accountID int64) error {
	return l.DB.Update(func(tx *bolt.Tx) error {
		_, err := l.resetOption(tx, optionName, accountID)
		return err
	})
}

// UpdateAccount ...
func (l *BillingLimiter) UpdateAccount(accountID int64, account BillingAccount) error {

	return l.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(billingDatabaseName))

		buff := bytes.NewBuffer([]byte{})
		if err := json.NewEncoder(buff).Encode(&account); err != nil {
			return err
		}

		err := b.Put([]byte(fmt.Sprintf("%v", accountID)), buff.Bytes())
		return err
	})

}

// DowngradeToDefaultPlan ...
func (l *BillingLimiter) DowngradeToDefaultPlan(accountID int64) error {
	l.Logger.Printf("plan for account(%v) expired, init account downgrade to default billing plan\n", accountID)
	defaultPlan := l.defaultAccountPlan
	defaultPlan.AccountID = accountID
	return l.ActualizePlanCounters(defaultPlan)
}

// ActualizePlanCounters ...
func (l *BillingLimiter) ActualizePlanCounters(p AccountBillingPlan) error {

	for i, opt := range p.Options {
		switch opt.Name {
		case "url_limit":
			cnt, err := l.UrlRepo.GetUserLinksCount(p.AccountID, p.Start, p.End)
			if err != nil {
				return err
			}
			topCount := p.Options[i].AsInt64()
			l.Logger.Printf("(account_id=%v) set billing value(%s) to %v, max=%v, value=%v", p.AccountID, opt.Name, topCount-int64(cnt), topCount, cnt)
			p.Options[i].Value = fmt.Sprintf("%v", topCount-int64(cnt))
		case "timedata_limit":
		case "users_limit":
		case "tags":
		case "tags_limit":
		case "groups":
		case "groups_limit":
		case "campaigns":
		case "campaigns_limit":
		case "rate_limit":
		default:
			return fmt.Errorf("billing option is not supported: %v", opt.Name)
		}
	}

	billingAccount := BillingAccount{
		Start:   p.Start,
		End:     p.End,
		Options: p.Options,
	}

	if err := l.UpdateAccount(p.AccountID, billingAccount); err != nil {
		return err
	}

	return nil
}

// LoadData fill billing database with actual billing information
func (l *BillingLimiter) LoadData() error {

	defaultPlan, err := l.Repo.GetDefaultPlan()
	if err != nil {
		return nil
	}

	curTime := utils.Now()

	l.defaultAccountPlan = AccountBillingPlan{
		BillingPlan: *defaultPlan,
		Start:       curTime,
		End:         curTime.Add(24 * 365 * 100 * time.Hour),
	}

	plans, err := l.Repo.GetAllUserBillingPlans(0)
	if err != nil {
		return errors.Wrap(err, "get account plans error:")
	}

	for _, p := range plans {

		// if current plan expired reset to default plan
		if !(p.Start.Before(curTime) && p.End.After(curTime)) {
			if err := l.DowngradeToDefaultPlan(p.AccountID); err != nil {
				return errors.Wrap(err, "downgrade to default plan error:")
			}
			continue
		}

		if err := l.ActualizePlanCounters(p); err != nil {
			return errors.Wrap(err, "actualize plan error:")
		}
	}

	return nil
}

var OptionNotFound = errors.New("option not found")

// GetOptionValue ...
func (l *BillingLimiter) GetOptionValue(optionName string, accountID int64) (*BillingOption, error) {
	var optionValue *BillingOption
	err := l.DB.View(func(tx *bolt.Tx) error {
		v, err := l.getOption(tx, optionName, accountID)
		if err == nil {
			optionValue = v
		}
		return err
	})
	return optionValue, err
}

// GetValue returns an actual billing value for provided option
func (l *BillingLimiter) getOption(tx *bolt.Tx, optionName string, accountID int64) (*BillingOption, error) {

	b := tx.Bucket([]byte(billingDatabaseName))

	v := b.Get([]byte(fmt.Sprintf("%v", accountID)))
	// cache miss only possible with plan reset
	if len(v) == 0 {
		l.Logger.Println("no billing found")
		option, err := l.resetOption(tx, optionName, accountID)
		if err != nil {
			return nil, err
		}
		return option, nil
	}

	buff := bytes.NewBuffer(v)
	var acc BillingAccount
	if err := json.NewDecoder(buff).Decode(&acc); err != nil {
		return nil, err
	}

	curTime := utils.Now()
	if !(curTime.After(acc.Start) && curTime.Before(acc.End)) {
		return nil, BillingAccountExpiredError
	}

	for _, option := range acc.Options {
		if option.Name == optionName {
			return &option, nil
		}
	}

	return nil, OptionNotFound
}

// CheckLimits ...
func (l *BillingLimiter) CheckLimits(optionName string, accountID int64) error {

	return l.DB.Update(func(tx *bolt.Tx) error {

		option, err := l.getOption(tx, optionName, accountID)
		if err == OptionNotFound {
			return LimitExceededError
		} else if err == BillingAccountExpiredError {
			if err := l.DowngradeToDefaultPlan(accountID); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		value := option.AsInt64()
		l.Logger.Printf("(account_id=%v) current billing value(%s) value: %v\n", accountID, optionName, value)
		if value <= 0 {
			return LimitExceededError
		}

		return nil
	})

}

func (l *BillingLimiter) getAccount(tx *bolt.Tx, accountID int64) (*BillingAccount, error) {

	b := tx.Bucket([]byte(billingDatabaseName))
	v := b.Get([]byte(fmt.Sprintf("%v", accountID)))

	if len(v) == 0 {
		return nil, LimitExceededError
	}

	buff := bytes.NewBuffer(v)

	var acc BillingAccount
	if err := json.NewDecoder(buff).Decode(&acc); err != nil {
		return nil, err
	}

	return &acc, nil
}

// UpdateOption ...
func (l *BillingLimiter) UpdateOption(tx *bolt.Tx, optionName string, accountID int64, f func(int64) int64) error {

	b := tx.Bucket([]byte(billingDatabaseName))
	account, err := l.getAccount(tx, accountID)
	if err != nil {
		return err
	}

	var update bool
	for i, option := range account.Options {
		if option.Name == optionName {
			value, _ := strconv.ParseInt(option.Value, 0, 64)
			updatedValue := f(value)
			if updatedValue < 0 {
				return errors.New("value is below zero")
			}
			l.Logger.Printf("(user=%v) changed billing value(%s) to %v", accountID, optionName, updatedValue)
			account.Options[i].Value = fmt.Sprintf("%v", updatedValue)
			update = true
			break
		}
	}

	if update {
		buff := bytes.NewBuffer([]byte{})
		if err := json.NewEncoder(buff).Encode(account); err != nil {
			return err
		}
		return b.Put([]byte(fmt.Sprintf("%v", accountID)), buff.Bytes())
	}

	return nil
}

// Reduce ...
func (l *BillingLimiter) Reduce(optionName string, accountID int64) error {
	return l.DB.Update(func(tx *bolt.Tx) error {
		return l.UpdateOption(tx, optionName, accountID, func(v int64) int64 {
			return v - 1
		})
	})
}

// Increase ...
func (l *BillingLimiter) Increase(optionName string, accountID int64) error {
	return l.DB.Update(func(tx *bolt.Tx) error {
		return l.UpdateOption(tx, optionName, accountID, func(v int64) int64 {
			return v + 1
		})
	})
}
