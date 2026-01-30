package instruments
import (
	"coefbot/types"
	"coefbot/opendota"
	"encoding/json"
	"os"
	"fmt"

)

var TournamentPlayers []types.Player

func LoadTournamentPlayers() {
    data, err := os.ReadFile("players.json")
    if err == nil {
        json.Unmarshal(data, &TournamentPlayers)
    }
}

func IsAdmin(userID int) bool {
    for _, adminID := range types.AdminIDs {
        if userID == adminID {
            return true
        }
    }
    return false
}

func InitHeroes() {
	data, err := os.ReadFile(types.HeroesFile)
	if err == nil {
		err = json.Unmarshal(data, &types.HeroMap)
		if err == nil {
			fmt.Println("Список героев успешно загружен из файла.")
			return
		}
	}

	fmt.Println("Файл не найден, загружаю данные из OpenDota API...")
	opendota.UpdateHeroesFromAPI()
}

func PlayersHeroesIdToNames(ps []types.Player) {
    for i := range ps {
        name, ok := types.HeroMap[ps[i].HeroID]
        if ok {
            ps[i].HeroName = name
        } else {
            ps[i].HeroName = "Unknown"
        }
    }
}

func SumFa(ps []float64) float64 {
	var sum float64
	for _, p := range ps {
		sum += p
	}
	return sum
}