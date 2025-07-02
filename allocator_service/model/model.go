package model

type RedisData struct {
	RoomID      string `json:"room_id"`
	CandidateID int    `json:"candidate_id"`
	RetryCount  int    `json:"retry_count"`
}
