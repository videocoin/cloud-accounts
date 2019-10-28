package datastore

import (
	"context"
	"fmt"

	opentracing "github.com/opentracing/opentracing-go"

	"github.com/jinzhu/gorm"
	v1 "github.com/videocoin/cloud-api/transfers/v1"
)

type TransferDatastore struct {
	db *gorm.DB
}

func NewTransferDatastore(db *gorm.DB) (*TransferDatastore, error) {
	db.AutoMigrate(&v1.Transfer{})
	return &TransferDatastore{db: db}, nil
}

func (ds *TransferDatastore) Get(ctx context.Context, id string) (*v1.Transfer, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Get")
	defer span.Finish()

	span.SetTag("id", id)

	transfer := new(v1.Transfer)
	if err := ds.db.Where("id = ?", id).First(&transfer).Error; err != nil {
		return nil, fmt.Errorf("failed to get transfer: %s", err)
	}

	return transfer, nil
}

func (ds *TransferDatastore) Update(ctx context.Context, transfer *v1.Transfer, updates map[string]interface{}) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "Update")
	defer span.Finish()

	span.SetTag("updates", updates)

	if err := ds.db.Model(transfer).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update transfer: %s", err)
	}

	return nil
}
