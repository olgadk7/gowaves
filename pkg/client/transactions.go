package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/wavesplatform/gowaves/pkg/crypto"
	"github.com/wavesplatform/gowaves/pkg/proto"
	"io"
	"net/http"
)

type Transactions struct {
	options Options
}

func NewTransactions(options Options) *Transactions {
	return &Transactions{
		options: options,
	}
}

// Get the number of unconfirmed transactions in the UTX pool
func (a *Transactions) UnconfirmedSize(ctx context.Context) (uint64, *Response, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/transactions/unconfirmed/size", a.options.BaseUrl),
		nil)
	if err != nil {
		return 0, nil, err
	}

	out := make(map[string]uint64)
	response, err := doHttp(ctx, a.options, req, &out)
	if err != nil {
		return 0, response, err
	}

	return out["size"], response, nil
}

type TransactionTypeVersion struct {
	Type    proto.TransactionType `json:"type"`
	Version byte                  `json:"version,omitempty"`
}

// Get transaction info
func (a *Transactions) Info(ctx context.Context, id crypto.Digest) (proto.Transaction, *Response, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/transactions/info/%s", a.options.BaseUrl, id.String()),
		nil)
	if err != nil {
		return nil, nil, err
	}

	buf := new(bytes.Buffer)
	response, err := doHttp(ctx, a.options, req, buf)
	if err != nil {
		return nil, response, err
	}

	b := buf.Bytes()

	tt := new(TransactionTypeVersion)
	err = json.NewDecoder(buf).Decode(tt)
	if err != nil {
		return nil, response, err
	}

	out, err := UnmarshalTransaction(tt, bytes.NewReader(b))
	if err != nil {
		return nil, response, &ParseError{Err: err}
	}

	return out, response, nil
}

func UnmarshalTransaction(t *TransactionTypeVersion, buf io.Reader) (proto.Transaction, error) {
	var out proto.Transaction
	switch t.Type {
	case proto.GenesisTransaction: // 1
		out = &proto.Genesis{}
	case proto.PaymentTransaction: // 2
		out = &proto.Payment{}
	case proto.IssueTransaction: // 3
		out = &proto.IssueV1{}
	case proto.TransferTransaction: // 4
		out = &proto.TransferV1{}
	case proto.ReissueTransaction: // 5
		out = &proto.ReissueV1{}
	case proto.BurnTransaction: // 6
		out = &proto.BurnV1{}
	case proto.ExchangeTransaction: // 7
		out = &proto.ExchangeV1{}
	case proto.LeaseTransaction: // 8
		out = &proto.LeaseV1{}
	case proto.LeaseCancelTransaction: // 9
		out = &proto.LeaseCancelV1{}
	case proto.CreateAliasTransaction: // 10
		out = &proto.CreateAliasV1{}
	case proto.MassTransferTransaction: // 11
		out = &proto.MassTransferV1{}
	case proto.DataTransaction: // 12
		out = &proto.DataV1{}
	case proto.SetScriptTransaction: // 13
		out = &proto.SetScriptV1{}
	case proto.SponsorshipTransaction: // 14
		out = &proto.SponsorshipV1{}
	}

	if out == nil {
		return nil, errors.Errorf("unknown transaction type %d version %d", t.Type, t.Version)
	}

	return out, json.NewDecoder(buf).Decode(out)
}
