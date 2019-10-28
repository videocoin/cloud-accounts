package datastore

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type Datastore struct {
	Account  *AccountDatastore
	Transfer *TransferDatastore
	User     *UserDatastore
}

func NewDatastore(uri string) (*Datastore, error) {
	ds := new(Datastore)

	db, err := gorm.Open("mysql", uri)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)

	accountsDs, err := NewAccountDatastore(db)
	if err != nil {
		return nil, err
	}

	ds.Account = accountsDs

	transfersDs, err := NewTransferDatastore(db)
	if err != nil {
		return nil, err
	}

	ds.Transfer = transfersDs

	userDs, err := NewUserDatastore(db)
	if err != nil {
		return nil, err
	}

	ds.User = userDs

	return ds, nil
}
