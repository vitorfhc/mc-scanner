package api

type PingAndListResponse struct {
	Address     string
	Players     PlayersInfo     `json:"players"`
	Description DescriptionInfo `json:"description"`
}

type DescriptionInfo struct {
	Text string `json:"text"`
}

type PlayersInfo struct {
	Max    int `json:"max"`
	Online int `json:"online"`
}
