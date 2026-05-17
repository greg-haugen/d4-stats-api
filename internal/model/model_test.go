package model_test

import (
	"encoding/json"
	"testing"

	"github.com/ghaugen/d4-stats-api/internal/model"
)

func TestPlayerStats_RoundTrip(t *testing.T) {
	original := model.PlayerStats{
		Username: "Hero#1234",
		Platform: "battlenet",
		Characters: []model.Character{
			{Name: "Draven", Class: "Barbarian", Level: 100, ParagonLevel: 200},
		},
		Season: model.SeasonData{SeasonNumber: 4, JourneyProgress: 75},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var got model.PlayerStats
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if got.Username != original.Username {
		t.Errorf("Username = %q, want %q", got.Username, original.Username)
	}
	if got.Platform != original.Platform {
		t.Errorf("Platform = %q, want %q", got.Platform, original.Platform)
	}
	if len(got.Characters) != 1 || got.Characters[0] != original.Characters[0] {
		t.Errorf("Characters = %+v, want %+v", got.Characters, original.Characters)
	}
	if got.Season != original.Season {
		t.Errorf("Season = %+v, want %+v", got.Season, original.Season)
	}
}

func TestPlayerStats_EmptyCharacters(t *testing.T) {
	ps := model.PlayerStats{
		Username:   "solo",
		Platform:   "steam",
		Characters: []model.Character{},
		Season:     model.SeasonData{},
	}

	data, err := json.Marshal(ps)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	raw, ok := m["characters"]
	if !ok {
		t.Fatal("characters field absent from JSON")
	}
	if string(raw) == "null" {
		t.Error("characters is null, want []")
	}
}

func TestPlayerStats_JSONKeys(t *testing.T) {
	ps := model.PlayerStats{
		Username:   "u",
		Platform:   "psn",
		Characters: []model.Character{},
		Season:     model.SeasonData{SeasonNumber: 1, JourneyProgress: 0},
	}

	data, _ := json.Marshal(ps)
	var m map[string]json.RawMessage
	json.Unmarshal(data, &m)

	for _, key := range []string{"username", "platform", "characters", "season"} {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON key %q missing", key)
		}
	}
}

func TestCharacter_JSONKeys(t *testing.T) {
	c := model.Character{Name: "n", Class: "Sorcerer", Level: 50, ParagonLevel: 10}
	data, _ := json.Marshal(c)
	var m map[string]json.RawMessage
	json.Unmarshal(data, &m)

	for _, key := range []string{"name", "class", "level", "paragon_level"} {
		if _, ok := m[key]; !ok {
			t.Errorf("Character JSON key %q missing", key)
		}
	}
}

func TestSeasonData_JSONKeys(t *testing.T) {
	s := model.SeasonData{SeasonNumber: 3, JourneyProgress: 50}
	data, _ := json.Marshal(s)
	var m map[string]json.RawMessage
	json.Unmarshal(data, &m)

	for _, key := range []string{"season_number", "journey_progress"} {
		if _, ok := m[key]; !ok {
			t.Errorf("SeasonData JSON key %q missing", key)
		}
	}
}
