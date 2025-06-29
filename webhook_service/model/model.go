package model

type Request struct {
	RoomID         string `json:"room_id"`
	CandidateAgent Agent  `json:"candidate_agent"`
}

type Agent struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type RedisData struct {
	RoomID      string `json:"room_id"`
	CandidateID int    `json:"candidate_id"`
}
