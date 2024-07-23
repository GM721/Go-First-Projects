package entities

type Order struct {
	ID       uint    `json:"orderID"`
	ClientID uint    `json:"clientID"`
	Amount   float32 `json:"amount"`
	Status   string  `json:"status"`
}
