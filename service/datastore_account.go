package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/jinzhu/gorm"
	"github.com/opentracing/opentracing-go"
	v1 "github.com/videocoin/cloud-api/accounts/v1"
	"github.com/videocoin/cloud-pkg/uuid4"
)

var (
	ErrAccountNotFound = errors.New("account is not found")
)

type AccountDatastore struct {
	db *gorm.DB
}

func NewAccountDatastore(db *gorm.DB) (*AccountDatastore, error) {
	db.AutoMigrate(&v1.Account{})
	return &AccountDatastore{db: db}, nil
}

func (ds *AccountDatastore) Create(ctx context.Context, userID string, passphrase string) (*v1.Account, error) {
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

	account := &v1.Account{
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

func (ds *AccountDatastore) Get(ctx context.Context, accountID string) (*v1.Account, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Get")
	defer span.Finish()

	span.SetTag("id", accountID)

	account := new(v1.Account)

	if err := ds.db.Where("id = ?", accountID).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAccountNotFound
		}

		return nil, fmt.Errorf("failed to get account by id: %s", err.Error())
	}

	return account, nil
}

func (ds *AccountDatastore) GetByOwner(ctx context.Context, userID string) (*v1.Account, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetByOwner")
	defer span.Finish()

	span.SetTag("owner_id", userID)

	account := new(v1.Account)

	if err := ds.db.Where("user_id = ?", userID).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAccountNotFound
		}

		return nil, fmt.Errorf("failed to get account by owner id: %s", err.Error())
	}

	return account, nil
}

func (ds *AccountDatastore) GetByAddress(ctx context.Context, address string) (*v1.Account, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "GetByAddress")
	defer span.Finish()

	span.SetTag("address", address)

	account := new(v1.Account)

	if err := ds.db.Where("address = ?", address).First(&account).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAccountNotFound
		}

		return nil, fmt.Errorf("failed to get account by id: %s", err.Error())
	}

	return account, nil
}

func (ds *AccountDatastore) List(ctx context.Context) ([]*v1.Account, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "List")
	defer span.Finish()

	accounts := []*v1.Account{}

	if err := ds.db.Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get accounts list: %s", err)
	}

	return accounts, nil
}

func (ds *AccountDatastore) UpdateBalance(ctx context.Context, account *v1.Account, balance float64) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "UpdateBalance")
	defer span.Finish()

	span.SetTag("account_id", account.Id)
	span.SetTag("balance", balance)

	time, err := ptypes.Timestamp(ptypes.TimestampNow())
	if err != nil {
		return err
	}

	account.Balance = balance
	account.UpdatedAt = &time

	updates := map[string]interface{}{
		"balance":   account.Balance,
		"updatedAt": account.UpdatedAt,
	}

	if err = ds.db.Model(account).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}
