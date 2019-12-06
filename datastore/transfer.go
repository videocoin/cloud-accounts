package datastore

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/jinzhu/gorm"
	opentracing "github.com/opentracing/opentracing-go"
	v1 "github.com/videocoin/cloud-api/transfers/v1"
	"github.com/videocoin/cloud-pkg/uuid4"
)

var (
	ErrTransferNotFound = errors.New("transfer is not found")
)

type TransferDatastore struct {
	db *gorm.DB
}

func NewTransferDatastore(db *gorm.DB) (*TransferDatastore, error) {
	db.AutoMigrate(&v1.Transfer{})
	return &TransferDatastore{db: db}, nil
}

func (ds *TransferDatastore) Create(ctx context.Context, userId, address string, amount []byte) (*v1.Transfer, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Create")
	defer span.Finish()

	tx := ds.db.Begin()

	id, err := uuid4.New()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	createdAt, err := ptypes.Timestamp(ptypes.TimestampNow())
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	ts, err := ptypes.TimestampProto(time.Now().Add(time.Minute * 10))
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	expiresAt, err := ptypes.Timestamp(ts)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	transfer := &v1.Transfer{
		Id:        id,
		UserId:    userId,
		Kind:      v1.TransferKindWithdraw,
		Pin:       encodeToString(6),
		ToAddress: address,
		Amount:    amount,
		CreatedAt: &createdAt,
		ExpiresAt: &expiresAt,
	}

	if err = tx.Create(transfer).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()

	return transfer, nil
}

func (ds *TransferDatastore) Get(ctx context.Context, id string) (*v1.Transfer, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "Get")
	defer span.Finish()

	span.SetTag("id", id)

	transfer := new(v1.Transfer)
	if err := ds.db.Where("id = ?", id).First(&transfer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrTransferNotFound
		}

		return nil, fmt.Errorf("failed to get transfer by id: %s", err.Error())
	}

	return transfer, nil
}

func (ds *TransferDatastore) ListByUser(ctx context.Context, userId string) ([]*v1.Transfer, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "ListByUser")
	defer span.Finish()

	span.SetTag("user_id", userId)

	transfers := []*v1.Transfer{}
	if err := ds.db.Where("user_id = ?", userId).Find(&transfers).Error; err != nil {
		return nil, fmt.Errorf("failed to get user transfers: %s", err)
	}

	return transfers, nil
}

func (ds *TransferDatastore) Delete(ctx context.Context, id string) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "Delete")
	defer span.Finish()

	span.SetTag("id", id)

	transfer := &v1.Transfer{
		Id: id,
	}
	if err := ds.db.Delete(transfer).Error; err != nil {
		return fmt.Errorf("failed to delete user transfer: %s", err)
	}

	return nil
}

func (ds *TransferDatastore) Update(ctx context.Context, transfer *v1.Transfer, updates map[string]interface{}) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "Update")
	span.SetTag("updates", updates)

	if err := ds.db.Model(transfer).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update transfer: %s", err)
	}

	return nil
}

func (ds *TransferDatastore) SetCompleted(ctx context.Context, transfer *v1.Transfer) error {
	return ds.Update(ctx, transfer, map[string]interface{}{
		"status": v1.TransferStatusCompleted,
	})
}

func (ds *TransferDatastore) SetFailed(ctx context.Context, transfer *v1.Transfer) error {
	return ds.Update(ctx, transfer, map[string]interface{}{
		"status": v1.TransferStatusFailed,
	})
}

func (ds *TransferDatastore) SetPendingNative(ctx context.Context, transfer *v1.Transfer, hash string) error {
	return ds.Update(ctx, transfer, map[string]interface{}{
		"status":       v1.TransferStatusPendingNative,
		"tx_native_id": hash,
	})
}

func (ds *TransferDatastore) SetExecutedNative(ctx context.Context, transfer *v1.Transfer) error {
	return ds.Update(ctx, transfer, map[string]interface{}{
		"status": v1.TransferStatusExecutedNative,
	})
}

func (ds *TransferDatastore) SetPendingErc20(ctx context.Context, transfer *v1.Transfer, hash string) error {
	return ds.Update(ctx, transfer, map[string]interface{}{
		"status":      v1.TransferStatusPendingErc,
		"tx_erc20_id": hash,
	})
}
