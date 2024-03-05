package proto

var (
	TransactionStatePending = 0
	TransactionStateSuccess = 1
	TransactionStateFailed  = 2

	TransactionActionUnknow   = 0
	TransactionActionDeposit  = 1
	TransactionActionWithdraw = 2
	TransactionActionTransfer = 3
)

type Transaction struct {
	ID        uint64 `json:"id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    int    `json:"amount"`
	State     int    `json:"state"`
	CreatedAt int64  `json:"created_at"`
}

type TransactionRequest struct {
	Action int    `json:"action"`
	From   string `json:"from"`
	To     string `json:"to"`
	Amount int    `json:"amount"`
}

type TransactionResponse struct {
	ID uint64 `json:"id"`
}
