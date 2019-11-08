package manager

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/copier"
	"github.com/opentracing/opentracing-go"
	v1 "github.com/videocoin/cloud-api/accounts/v1"
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

	balance := new(big.Int)
	balance, _ = balance.SetString(string(account.Balance), 10)

	return &v1.AccountProfile{
		Address: account.Address,
		Balance: balance.String(),
	}, nil
}

func (m *Manager) ListAccounts(ctx context.Context) ([]*v1.AccountProfile, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.ListAccounts")
	defer span.Finish()

	accounts, err := m.ds.Account.List(ctx)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	profiles := make([]*v1.AccountProfile, len(accounts))
	for i, account := range accounts {
		profiles[i].Address = account.Address

		balance := new(big.Int)
		balance, _ = balance.SetString(string(account.Balance), 10)
		profiles[i].Balance = balance.String()
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

	balance := new(big.Int)
	balance, _ = balance.SetString(string(account.Balance), 10)

	return &v1.AccountProfile{
		Address: account.Address,
		Balance: balance.String(),
	}, nil
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

	balance := new(big.Int)
	balance, _ = balance.SetString(string(account.Balance), 10)

	return &v1.AccountProfile{
		Address: account.Address,
		Balance: balance.String(),
	}, nil
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

	balance := new(big.Int)
	balance, _ = balance.SetString(string(account.Balance), 10)

	return &v1.AccountProfile{
		Address: account.Address,
		Balance: balance.String(),
	}, nil
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

func (m *Manager) refreshBalance(ctx context.Context, account *v1.Account) error {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.refreshBalance")
	defer span.Finish()

	address := common.HexToAddress(account.Address)
	balance, err := m.vdc.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return err
	}

	if err = m.ds.Account.UpdateBalance(ctx, account, balance.String()); err != nil {
		tracer.SpanLogError(span, err)
		return err
	}

	return nil
}
