package manager

import (
	"context"

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

	return &v1.AccountProfile{
		Address: account.Address,
		Balance: "0",
	}, nil
}

func (m *Manager) GetAccounts(ctx context.Context) (*v1.Accounts, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.GetAccounts")
	defer span.Finish()

	accounts, err := m.ds.Account.List(ctx)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	profiles := &v1.Accounts{Items: []*v1.AccountProfile{}}
	for _, account := range accounts {
		profiles.Items = append(profiles.Items, &v1.AccountProfile{
			Address: account.Address,
			Balance: "0",
		})
	}

	return profiles, nil
}

func (m *Manager) GetAccountByID(ctx context.Context, id string) (*v1.AccountProfile, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.GetAccountByID")
	defer span.Finish()

	account, err := m.ds.Account.Get(ctx, id)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	return &v1.AccountProfile{
		Address: account.Address,
		Balance: account.BalanceWei,
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

	return &v1.AccountProfile{
		Address: account.Address,
		Balance: account.BalanceWei,
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

	return &v1.AccountProfile{
		Address: account.Address,
		Balance: account.BalanceWei,
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

func (m *Manager) GetAccountKeys(ctx context.Context) ([]*v1.AccountKey, error) {
	span, _ := opentracing.StartSpanFromContext(ctx, "manager.GetAccountKeys")
	defer span.Finish()

	accounts, err := m.ds.Account.List(ctx)
	if err != nil {
		tracer.SpanLogError(span, err)
		return nil, err
	}

	keys := make([]*v1.AccountKey, len(accounts))
	for i, account := range accounts {
		keys[i] = &v1.AccountKey{
			Key: account.Key,
		}
	}

	return keys, nil
}
