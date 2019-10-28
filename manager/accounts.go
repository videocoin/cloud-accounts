package manager

import (
	"context"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/copier"
	"github.com/opentracing/opentracing-go"
	v1 "github.com/videocoin/cloud-api/accounts/v1"
	transfersv1 "github.com/videocoin/cloud-api/transfers/v1"
	"github.com/videocoin/cloud-pkg/ethutils"
	"github.com/videocoin/cloud-pkg/tracer"
)

func (m *Manager) CreateAccount(ctx context.Context, req *v1.AccountRequest) (*v1.AccountProfile, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.CreateAccount")
	defer span.Finish()

	account, err := m.ds.Account.Create(ctx, req.OwnerId, m.clientSecret)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	profile := new(v1.AccountProfile)
	err = copier.Copy(profile, account)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	return profile, nil
}

func (m *Manager) ListAccounts(ctx context.Context) ([]*v1.AccountProfile, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.ListAccounts")
	defer span.Finish()

	accounts, err := m.ds.Account.List(ctx)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	profiles := make([]*v1.AccountProfile, 0)
	err = copier.Copy(&profiles, &accounts)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	return profiles, nil
}

func (m *Manager) GetAccountById(ctx context.Context, id string) (*v1.AccountProfile, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.GetAccountById")
	defer span.Finish()

	account, err := m.ds.Account.Get(ctx, id)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	err = m.refreshBalance(ctx, account)
	if err != nil {
		m.logger.WithError(err).Errorf("failed to refresh account %s balance", account.Id)
		tracer.SpanLogError(span, err)
	}

	profile := new(v1.AccountProfile)
	err = copier.Copy(profile, account)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	return profile, nil
}

func (m *Manager) GetAccountByOwner(ctx context.Context, ownerID string) (*v1.AccountProfile, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.GetAccountByOwner")
	defer span.Finish()

	account, err := m.ds.Account.GetByOwner(ctx, ownerID)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	err = m.refreshBalance(ctx, account)
	if err != nil {
		m.logger.WithError(err).Errorf("failed to refresh account %s balance", account.Id)
		tracer.SpanLogError(span, err)
	}

	profile := new(v1.AccountProfile)
	err = copier.Copy(profile, account)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	return profile, nil
}

func (m *Manager) GetAccountByAddress(ctx context.Context, address string) (*v1.AccountProfile, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.GetAccountByAddress")
	defer span.Finish()

	account, err := m.ds.Account.GetByAddress(ctx, address)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	err = m.refreshBalance(ctx, account)
	if err != nil {
		m.logger.WithError(err).Errorf("failed to refresh account %s balance", account.Id)
		tracer.SpanLogError(span, err)
	}

	profile := new(v1.AccountProfile)
	err = copier.Copy(profile, account)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	return profile, nil
}

func (m *Manager) GetAccountKey(ctx context.Context, ownerID string) (*v1.AccountKey, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.GetAccountKey")
	defer span.Finish()

	account, err := m.ds.Account.GetByOwner(ctx, ownerID)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	key := new(v1.AccountKey)
	err = copier.Copy(key, account)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	return key, nil
}

func (m *Manager) Withdraw(key *v1.AccountKey, req *v1.WithdrawRequest) {
	go m.withdraw(context.Background(), key, req)
}

func (m *Manager) withdraw(ctx context.Context, key *v1.AccountKey, req *v1.WithdrawRequest) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.withdraw")
	defer span.Finish()

	isSuccessfull := false
	failReason := "Unknown reason"

	transfer, err := m.ds.Transfer.Get(ctx, req.TransferId)
	if err != nil {
		m.logger.WithError(err).Error("failed to query transfer")
		return
	}

	defer func() {
		user, err := m.ds.User.Get(ctx, transfer.UserId)
		if err != nil {
			m.logger.WithError(err).Error("failed to query user")
			return
		}

		if !isSuccessfull {
			if err := m.ds.Transfer.Update(ctx, transfer, map[string]interface{}{"status": transfersv1.TransferStatusFailed}); err != nil {
				m.logger.WithError(err).Error("failed to update transfer")
			}
			if err := m.nc.SendWithdrawFailed(ctx, user, transfer, failReason); err != nil {
				m.logger.WithError(err).Error("failed to send withdraw failed email")
			}

		} else {
			if err := m.ds.Transfer.Update(ctx, transfer, map[string]interface{}{"status": transfersv1.TransferStatusCompleted}); err != nil {
				m.logger.WithError(err).Error("failed to update transfer")
			}
			if err := m.nc.SendWithdrawSucceeded(ctx, user, transfer); err != nil {
				m.logger.WithError(err).Error("failed to send withdraw succeeded email")
			}
		}
	}()

	transferAmount := ethutils.EthToWei(transfer.Amount)

	m.logger.WithField("to_address", transfer.ToAddress).WithField("amount", transfer.Amount).Info("starting withdraw procedure")

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

func (m *Manager) refreshBalance(ctx context.Context, account *v1.Account) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.refreshBalance")
	defer span.Finish()

	address := common.HexToAddress(account.Address)
	weiAmount, err := m.vdc.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return err
	}

	vdcAmount := ethutils.WeiToEth(weiAmount)
	if err = m.ds.Account.UpdateBalance(ctx, account, vdcAmount); err != nil {
		tracer.SpanLogError(span, err)
		return err
	}

	account.Balance = vdcAmount

	return nil
}
