package mcscanner

type Options struct {
	InputChan   chan string
	ResultsChan chan *PingAndListResponse
	Timeout     int
	MaxJobs     int
}

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
