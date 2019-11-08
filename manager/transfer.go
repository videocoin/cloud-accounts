package manager

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/copier"
	"github.com/opentracing/opentracing-go"
	v1 "github.com/videocoin/cloud-api/accounts/v1"
	transfersv1 "github.com/videocoin/cloud-api/transfers/v1"
	"github.com/videocoin/cloud-pkg/tracer"
)

func (m *Manager) CreateTransfer(ctx context.Context, req *v1.CreateTransferRequest) (*v1.TransferResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.CreateTransfer")
	defer span.Finish()

	transfer, err := m.ds.Transfer.Create(ctx, req.UserId, req.ToAddress, req.Amount)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	resp := new(v1.TransferResponse)
	err = copier.Copy(resp, transfer)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	return resp, nil
}

func (m *Manager) GetTransfer(ctx context.Context, req *v1.TransferRequest) (*v1.TransferResponse, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.GetTransfer")
	defer span.Finish()

	transfer, err := m.ds.Transfer.Get(ctx, req.Id)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	resp := new(v1.TransferResponse)
	err = copier.Copy(resp, transfer)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	return resp, nil
}

func (m *Manager) ExecuteTransfer(ctx context.Context, key *v1.AccountKey, req *v1.ExecuteTransferRequest) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.ExecuteTransfer")
	defer span.Finish()
	go m.executeTransfer(context.Background(), key, req)
}

func (m *Manager) executeTransfer(ctx context.Context, key *v1.AccountKey, req *v1.ExecuteTransferRequest) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.executeTransfer")
	defer span.Finish()

	isSuccessfull := false
	failReason := "Unknown reason"

	transfer, err := m.ds.Transfer.Get(ctx, req.Id)
	if err != nil {
		m.logger.WithError(err).Error("failed to query transfer")
		return
	}

	transferAmount := new(big.Int)
	transferAmount, _ = transferAmount.SetString(string(transfer.Amount), 10)

	m.logger.WithField("to_address", transfer.ToAddress).WithField("amount", transferAmount.Uint64()).Info("starting withdraw procedure")

	defer func() {
		if !isSuccessfull {
			if err := m.ds.Transfer.Update(ctx, transfer, map[string]interface{}{"status": transfersv1.TransferStatusFailed}); err != nil {
				m.logger.WithError(err).Error("failed to update transfer")
			}

			if err := m.nc.SendWithdrawFailed(ctx, req.UserEmail, transfer, failReason); err != nil {
				m.logger.WithError(err).Error("failed to send withdraw failed email")
			}

		} else {
			if err := m.ds.Transfer.Update(ctx, transfer, map[string]interface{}{"status": transfersv1.TransferStatusCompleted}); err != nil {
				m.logger.WithError(err).Error("failed to update transfer")
			}
			if err := m.nc.SendWithdrawSucceeded(ctx, req.UserEmail, transfer); err != nil {
				m.logger.WithError(err).Error("failed to send withdraw succeeded email")
			}
		}
	}()

	userKey, err := keystore.DecryptKey([]byte(key.Key), m.clientSecret)
	if err != nil {
		m.logger.Error(err)
		return
	}

	if err = checkUserBalance(m.vdc, userKey.Address, transferAmount); err != nil {
		failReason = "Unsufficient user balance"
		m.logger.Error(err)
		return
	}

	if err = checkBankBalance(m.eth, m.bankKey.Address, common.HexToAddress(m.tokenAddr), transferAmount); err != nil {
		failReason = "Unsufficient bank balance"
		m.logger.Error(err)
		return
	}

	tx, err := execNativeTransaction(m.vdc, userKey, m.bankKey.Address, transferAmount)
	if err != nil {
		failReason = "Native transaction failed"
		m.logger.Error(err)
		return
	}

	if err := m.ds.Transfer.Update(ctx, transfer, map[string]interface{}{
		"status":       transfersv1.TransferStatusExecutedNative,
		"tx_native_id": tx.Hash().String()}); err != nil {
		m.logger.WithError(err).Error("failed to update transfer")
	}

	tx, err = execErc20Transaction(m.eth, m.bankKey, common.HexToAddress(transfer.ToAddress), common.HexToAddress(m.tokenAddr), transferAmount)
	if err != nil {
		failReason = "Erc20 transaction failed"
		m.logger.Error(err)
		return
	}

	if err := m.ds.Transfer.Update(ctx, transfer, map[string]interface{}{
		"status":      transfersv1.TransferStatusExecutedNative,
		"tx_erc20_id": tx.Hash().String()}); err != nil {
		m.logger.WithError(err).Error("failed to update transfer")
	}

	isSuccessfull = true
}
