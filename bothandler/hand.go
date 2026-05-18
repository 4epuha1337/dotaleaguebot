package bothandler

import (
	"encoding/json"
	"fmt"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	fal "coefbot/facoef"
	"coefbot/instruments"
	"coefbot/types"
)

// awaitingJSON tracks which chat IDs are waiting for manual JSON input.
var (
	awaitingJSON   = make(map[int64]bool)
	awaitingJSONMu sync.Mutex
)

func setAwaiting(chatID int64, val bool) {
	awaitingJSONMu.Lock()
	defer awaitingJSONMu.Unlock()
	if val {
		awaitingJSON[chatID] = true
	} else {
		delete(awaitingJSON, chatID)
	}
}

func isAwaiting(chatID int64) bool {
	awaitingJSONMu.Lock()
	defer awaitingJSONMu.Unlock()
	return awaitingJSON[chatID]
}

// HandleHandJSON parses raw match JSON sent by the user and replies
// with the full /details-style breakdown.
func HandleHandJSON(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	chatID := update.Message.Chat.ID

	// Accept JSON from plain text or a document (.json file).
	var raw []byte
	if update.Message.Document != nil {
		// Download the file Telegram is hosting.
		file, err := bot.GetFile(tgbotapi.FileConfig{FileID: update.Message.Document.FileID})
		if err != nil {
			bot.Send(tgbotapi.NewMessage(chatID, "Не удалось скачать файл с серверов Telegram."))
			return
		}
		link := file.Link(bot.Token)
		import_http_resp, err2 := fetchURL(link)
		if err2 != nil {
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Ошибка при скачивании файла: %v", err2)))
			return
		}
		raw = import_http_resp
	} else {
		raw = []byte(update.Message.Text)
	}

	var match types.Match
	if err := json.Unmarshal(raw, &match); err != nil {
		bot.Send(tgbotapi.NewMessage(chatID,
			fmt.Sprintf("Не удалось распарсить JSON: %v\n\nПроверь формат и попробуй снова, или введи /hand заново.", err)))
		return
	}

	// Same validation as /analyze and /details.
	if len(match.Players) != 10 {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка: в JSON должно быть ровно 10 игроков (поле \"players\")."))
		return
	}
	if len(match.Players[0].GoldTimes) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка: отсутствуют данные gold_t. Убедись, что матч запаршен на OpenDota."))
		return
	}
	if len(match.GoldAdv) == 0 {
		bot.Send(tgbotapi.NewMessage(chatID, "Ошибка: отсутствует поле radiant_gold_adv. Убедись, что матч запаршен на OpenDota."))
		return
	}

	winnerText := "🏆*Победитель*: Radiant\n"
	if !match.IsRadWin {
		winnerText = "🏆*Победитель*: Dire\n"
	}

	instruments.PlayersHeroesIdToNames(match.Players)
	FaR, tSumR, macroR, domR, playersR := fal.FaCoef(&match, match.Players[:5], match.TowersD, 1)
	FaD, tSumD, macroD, domD, playersD := fal.FaCoef(&match, match.Players[5:], match.TowersR, -1)

	text := fmt.Sprintf(`%s
*Коэффициент Фа для Radiant* - %.4f,
*Коэффициент Фа для Dire* - %.4f
 
*Командная сумма Radiant* - %.4f
*Командная сумма Dire* - %.4f
 
*Макро коэффициент Radiant* - %.4f
*Макро коэффициент Dire* - %.4f
 
*Индекс доминации Radiant* - %.4f
*Индекс доминации Dire* - %.4f
				
*Личный рейтинг игроков Radiant:* 
%d. *%s* - %.4f
%d. *%s* - %.4f
%d. *%s* - %.4f
%d. *%s* - %.4f
%d. *%s* - %.4f
				
*Личный рейтинг игроков Dire:*
%d. *%s* - %.4f
%d. *%s* - %.4f
%d. *%s* - %.4f
%d. *%s* - %.4f
%d. *%s* - %.4f`,
		winnerText, FaR, FaD, tSumR, tSumD, macroR, macroD, domR, domD,
		1, match.Players[0].HeroName, playersR[0],
		2, match.Players[1].HeroName, playersR[1],
		3, match.Players[2].HeroName, playersR[2],
		4, match.Players[3].HeroName, playersR[3],
		5, match.Players[4].HeroName, playersR[4],
		1, match.Players[5].HeroName, playersD[0],
		2, match.Players[6].HeroName, playersD[1],
		3, match.Players[7].HeroName, playersD[2],
		4, match.Players[8].HeroName, playersD[3],
		5, match.Players[9].HeroName, playersD[4],
	)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	bot.Send(msg)
}
