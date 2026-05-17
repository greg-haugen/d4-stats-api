package model

type PlayerStats struct {
	Username   string      `json:"username"`
	Platform   string      `json:"platform"`
	Characters []Character `json:"characters"`
	Season     SeasonData  `json:"season"`
}

type Character struct {
	Name         string `json:"name"`
	Class        string `json:"class"`
	Level        int    `json:"level"`
	ParagonLevel int    `json:"paragon_level"`
}

type SeasonData struct {
	SeasonNumber    int `json:"season_number"`
	JourneyProgress int `json:"journey_progress"`
}
