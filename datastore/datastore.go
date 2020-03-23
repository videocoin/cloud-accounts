package datastore

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //nolint
)

type Datastore struct {
	Account  *AccountDatastore
	Transfer *TransferDatastore
}

func NewDatastore(uri string) (*Datastore, error) {
	ds := new(Datastore)

	db, err := gorm.Open("mysql", uri)
	if err != nil {
		return nil, err
	}

	//db.LogMode(true)

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

	return ds, nil
}
