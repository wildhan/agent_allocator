package model

type RedisData struct {
	RoomID      string `json:"room_id"`
	CandidateID int    `json:"candidate_id"`
	RetryCount  int    `json:"retry_count"`
}

type ResponseGetAgentByID struct {
	Data []Agent `json:"data"`
}

type Agent struct {
	ID                   int    `json:"id"`
	Name                 string `json:"name"`
	CurrentCustomerCount int    `json:"current_customer_count"`
}

type ResponseGetAvailableAgent struct {
	Data data `json:"data"`
}

type data struct {
	Agents []Agent `json:"agents"`
}
