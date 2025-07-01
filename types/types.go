package types

type Cart struct {
	ID        int    `json:"id"`
	UserID    int    `json:"userId"`
	ProductID int    `json:"productId"`
	Quantity  int    `json:"quantity"`
	Status    string `json:"status"`
}

type UserCartResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    []Cart `json:"data"`
}

type CartResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    Cart   `json:"data"`
}

type Call struct {
	ProductId int `json:"product_id"`
	Quantity  int `json:"quantity"`
}
