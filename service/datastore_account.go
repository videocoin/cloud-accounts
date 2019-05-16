package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/AlekSi/pointer"
	v1 "github.com/VideoCoin/cloud-api/accounts/v1"
	"github.com/VideoCoin/cloud-pkg/uuid4"
	"github.com/jinzhu/gorm"
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

func (ds *AccountDatastore) Create(userID string, passphrase string) (*v1.Account, error) {
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

func (ds *AccountDatastore) Get(userID string) (*v1.Account, error) {
	account := new(v1.Account)

	err := ds.db.Where("user_id = ?", userID).First(&account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAccountNotFound
		}

		return nil, fmt.Errorf("failed to get account by id: %s", err.Error())
	}

	return account, nil
}

func (ds *AccountDatastore) GetByAddress(address string) (*v1.Account, error) {
	account := new(v1.Account)

	err := ds.db.Where("address = ?", address).First(&account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrAccountNotFound
		}

		return nil, fmt.Errorf("failed to get account by id: %s", err.Error())
	}

	return account, nil
}

func (ds *AccountDatastore) List() ([]*v1.Account, error) {
	accounts := []*v1.Account{}

	err := ds.db.Find(&accounts).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts list: %s", err)
	}

	return accounts, nil
}

func (ds *AccountDatastore) UpdateBalance(account *v1.Account, balance float64) error {
	account.Balance = balance
	account.UpdatedAt = pointer.ToTime(time.Now())

	updates := map[string]interface{}{
		"balance":   account.Balance,
		"updatedAt": account.UpdatedAt,
	}

	err := ds.db.Model(account).Updates(updates).Error
	if err != nil {
		return err
	}

	return nil
}
