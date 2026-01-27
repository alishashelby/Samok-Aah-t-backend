package context

type ctxKey string

const (
	CorrelationID ctxKey = "correlation_id"
	Tx            ctxKey = "tx"
)

func (ctx ctxKey) String() string {
	return string(ctx)
}
