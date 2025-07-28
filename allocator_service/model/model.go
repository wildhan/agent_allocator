package model

type RedisData struct {
	RoomID      string `json:"room_id"`
	CandidateID int    `json:"candidate_id"`
	RetryCount  int    `json:"retry_count"`
}

type ResponseGetAgent struct {
	Data []Agent `json:"data"`
}

type Agent struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	CurrentCustomerCount int    `json:"current_customer_count"`
}
