package proto

type User struct {
	Account   string `json:"account"`
	Password  string `json:"password"`
	Name      string `json:"name"`
	Balance   int    `json:"balance"`
	Nonce     string `json:"nonce"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type UserToken struct {
	Account   string `json:"account"`
	Nonce     string `json:"nonce"`
	ExpireAt  int64  `json:"expire_at"`
	CreatedAt int64  `json:"created_at"`
}

type GetTokenRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type CreateAccountRequest struct {
	Password string `json:"password"`
	Name     string `json:"name"`
	Balance  int    `json:"balance"`
}
