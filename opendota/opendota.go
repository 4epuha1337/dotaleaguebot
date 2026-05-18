package opendota

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"coefbot/types"
)

func GetMatch(matchID int64) (*types.Match, error) {
	url := fmt.Sprintf("https://api.opendota.com/api/matches/%d", matchID)
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var match types.Match
	if err := json.Unmarshal(body, &match); err != nil {
		return nil, err
	}
	return &match, nil
}

func UpdateHeroesFromAPI() {
	resp, err := http.Get("https://api.opendota.com/api/heroes")
	if err != nil {
		fmt.Printf("Ошибка при запросе к API: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var heroes []struct {
		ID   int    `json:"id"`
		Name string `json:"localized_name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&heroes); err != nil {
		fmt.Printf("Ошибка декодирования API: %v\n", err)
		return
	}

	for _, h := range heroes {
		types.HeroMap[h.ID] = h.Name
	}

	fileData, err := json.MarshalIndent(types.HeroMap, "", "  ")
	if err != nil {
		fmt.Printf("Ошибка сериализации в JSON: %v\n", err)
		return
	}

	err = os.WriteFile(types.HeroesFile, fileData, 0644)
	if err != nil {
		fmt.Printf("Ошибка записи в файл: %v\n", err)
	} else {
		fmt.Println("Данные героев сохранены в heroes.json")
	}
}
