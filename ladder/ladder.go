package ladder

import (
	"coefbot/types"
	
	"os"
	"strings"
	"sort"
	"strconv"
	"encoding/json"
	"fmt"
	"math"
)

func ptsStandardDeviation(p []types.Player) float64 {
    if len(p) == 0 {return 0}
    var sum, mean, sd float64
    for _, n := range p {
       sum += float64(n.DotaMMR)
    }
    mean = sum / float64(len(p))
    for _, n := range p {
       sd += math.Pow(float64(n.DotaMMR) - mean, 2)
    }
    sd = math.Sqrt(sd / float64(len(p)))
    return sd
}

func LoadLadder() ([]types.Player, error) {
    data, err := os.ReadFile("players.json")
    if err != nil {
        return nil, err
    }
    var players []types.Player
    err = json.Unmarshal(data, &players)
    return players, err
}

func SaveLadder(players []types.Player) error {
    data, err := json.MarshalIndent(players, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile("players.json", data, 0644)
}

func sumPTS(players []types.Player) (int) {
	var sum int
	for _, p := range players {
		sum += p.DotaMMR
	}
	return sum
}

func IsMatchProcessed(matchID int64) bool {
    data, err := os.ReadFile("history.txt")
    if err != nil { return false }
    return strings.Contains(string(data), strconv.FormatInt(matchID, 10))
}

func MarkMatchProcessed(matchID int64) {
    f, _ := os.OpenFile("history.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    defer f.Close()
    f.WriteString(strconv.FormatInt(matchID, 10) + "\n")
}

func CalculateMatchPTS(allPlayers []types.Player, players1 []types.Player, players2 []types.Player, isWinner bool) int {
	diff := sumPTS(players1) - sumPTS(players2)
	absDiff := math.Abs(float64(diff))
	if absDiff < 100 {
		return 125
	}
	sd := ptsStandardDeviation(allPlayers)
	threshold := sd * 2
	if threshold <= 100 {
		threshold = 200
	}
	factor := (absDiff - 100) / (threshold - 100)
	if factor > 1 {
		factor = 1
	}
	isLowWin := (diff < 0 && isWinner) || (diff > 0 && ! isWinner)

	if isLowWin {
		return int(math.Floor(125 + factor * 75))
	}
	return int(math.Floor(125 - factor * 75))
}

func GetTopFormatted(players []types.Player) string {
    temp := make([]types.Player, len(players))
    copy(temp, players)

    sort.Slice(temp, func(i, j int) bool {
        return temp[i].TournamentPTS > temp[j].TournamentPTS
    })

    var b strings.Builder
    b.WriteString("🏆 *Таблица лидеров турнира:*\n\n")

    for i, p := range temp {
        if i >= 10 { break }
        
        icon := "👤"
        if i == 0 { icon = "🥇" }
        if i == 1 { icon = "🥈" }
        if i == 2 { icon = "🥉" }

        b.WriteString(fmt.Sprintf("%s %d. *%s* — %.1f pts\n└ _avg impact: %.2f_ | _игр: %d_\n\n", 
            icon, i+1, p.Nickname, p.TournamentPTS, p.AverageFa, p.MatchesPlayed))
    }
    
    return b.String()
}

func ApplyMatchResults(allPlayers []types.Player, team []types.Player, pool int, teamSumFa float64, playerFaValues []float64) []types.Player {
	for i := range team {
		idx := -1
		for j := range allPlayers {
			if allPlayers[j].AccountID == team[i].AccountID {
				idx = j
				break
			}
		}

		if idx == -1 {
			newPlayer := types.Player{
				AccountID: team[i].AccountID,
				Nickname:  team[i].Nickname,
				DotaMMR:   team[i].DotaMMR, 
				Role:      team[i].Role,
				TournamentPTS: 1000,
			}
			allPlayers = append(allPlayers, newPlayer)
			idx = len(allPlayers) - 1
		}

		if idx != -1 {
			share := playerFaValues[i] / teamSumFa
			gain := float64(pool) * share
			
			if pool < 0 {
				if gain > -5 {gain = -5}
				if gain < -40 {gain = -40}
			} else {
				if gain > 40 { gain = 40 }
				if gain < 5  { gain = 5  }
			}

			allPlayers[idx].TournamentPTS += gain
			allPlayers[idx].MatchesPlayed++

			n := float64(allPlayers[idx].MatchesPlayed)
			allPlayers[idx].AverageFa = (allPlayers[idx].AverageFa*(n-1) + playerFaValues[i]) / n
		}
	}
	return allPlayers
}