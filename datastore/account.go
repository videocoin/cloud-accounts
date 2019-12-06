package datastore

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/jinzhu/gorm"
	"github.com/opentracing/opentracing-go"
	"github.com/videocoin/cloud-pkg/uuid4"
)

var (
	ErrAccountNotFound = errors.New("account is not found")
)

type AccountDatastore struct {
	db *gorm.DB
}

func NewAccountDatastore(db *gorm.DB) (*AccountDatastore, error) {
	return &AccountDatastore{db: db}, nil
}

type Account struct {
	Id         string     `gorm:"type:varchar(36);PRIMARY_KEY"`
	UserId     string     `gorm:"type:varchar(255);DEFAULT:null"`
	Address    string     `gorm:"type:varchar(42);DEFAULT:null"`
	Key        string     `gorm:"type:varchar(42);DEFAULT:null"`
	UpdatedAt  *time.Time `gorm:"type:timestamp NULL;DEFAULT:null"`
	Balance    string     `gorm:"type:double;DEFAULT:null"`
	BalanceWei string     `gorm:"type:varchar(255);DEFAULT:null"`
}

func (ds *AccountDatastore) Create(ctx context.Context, userID string, passphrase string) (*Account, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Create")
	defer span.Finish()

	span.SetTag("owner_id", userID)

	tx := ds.db.Begin()

	id, err := uuid4.New()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	key, err := generateKey(passphrase)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	account := &Account{
		Id:      id,
		UserId:  userID,
		Address: key.Address,
		Key:     key.KeyFile,
	}

	err = tx.Create(account).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return account, nil
}

func (ds *AccountDatastore) Get(ctx context.Context, id string) (*Account, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Get")
	defer span.Finish()

	span.SetTag("id", id)

	account := new(Account)
	if err := ds.db.Where("id = ?", id).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAccountNotFound
		}

		return nil, fmt.Errorf("failed to get account by id: %s", err.Error())
	}

	return account, nil
}

func (ds *AccountDatastore) GetByOwner(ctx context.Context, userID string) (*Account, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetByOwner")
	defer span.Finish()

	span.SetTag("owner_id", userID)

	account := new(Account)

	if err := ds.db.Where("user_id = ?", userID).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAccountNotFound
		}

		return nil, fmt.Errorf("failed to get account by owner id: %s", err.Error())
	}

	return account, nil
}

func (ds *AccountDatastore) GetByAddress(ctx context.Context, address string) (*Account, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetByAddress")
	defer span.Finish()

	span.SetTag("address", address)

	account := new(Account)

	if err := ds.db.Where("address = ?", address).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAccountNotFound
		}

		return nil, fmt.Errorf("failed to get account by id: %s", err.Error())
	}

	return account, nil
}

func (ds *AccountDatastore) List(ctx context.Context) ([]*Account, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "List")
	defer span.Finish()

	accounts := []*Account{}

	if err := ds.db.Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get accounts list: %s", err)
	}

	return accounts, nil
}

func (ds *AccountDatastore) Update(ctx context.Context, account *Account, updates map[string]interface{}) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "Update")
	span.SetTag("updates", updates)

	if err := ds.db.Model(account).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update account: %s", err)
	}

	return nil
}

func (ds *AccountDatastore) SetBalance(ctx context.Context, account *Account, balance *big.Int) error {
	time, err := ptypes.Timestamp(ptypes.TimestampNow())
	if err != nil {
		return err
	}

	return ds.Update(ctx, account, map[string]interface{}{
		"balance_wei": balance.String(),
		"updated_at":  &time,
	})
}

func (ds *AccountDatastore) Lock(ctx context.Context, account *Account) error {
	return ds.Update(ctx, account, map[string]interface{}{
		"is_locked": true,
	})
}

func (ds *AccountDatastore) Unlock(ctx context.Context, account *Account) error {
	return ds.Update(ctx, account, map[string]interface{}{
		"is_locked": false,
	})
}
