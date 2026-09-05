package pool

import "context"

type TxRunner struct {
	beginner TxBeginner
}

func NewTxRunner(beginner TxBeginner) *TxRunner {
	return &TxRunner{beginner: beginner}
}

func (r *TxRunner) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return WithinTx(ctx, r.beginner.Begin, fn)
}
