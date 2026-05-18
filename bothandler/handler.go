package bothandler

import (
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

	fal "coefbot/facoef"
	"coefbot/instruments"
	"coefbot/ladder"
	"coefbot/opendota"
)

func MessageHandler(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	winnerText := "🏆*Победитель*: Radiant\n"
	if update.Message != nil {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		if isAwaiting(update.Message.Chat.ID) {
			if update.Message.Text == "cancel" {
				awaitingJSON[update.Message.Chat.ID] = false
			} else {
				HandleHandJSON(bot, update)
				awaitingJSON[update.Message.Chat.ID] = false
			}
		} else {
			switch update.Message.Command() {
			case "start":
				msgText := `Привет! Данный бот предназначен исключительно для расчета турнирного коэффициента ANTON.

📥 Как пользоваться: 
/analyze <match_id> - Коэффициент результативности ANTON для обеих команд.
/details <match_id> - Полный набор всех данных, вычисленных во время работы: Коэффициент ANTON, личный рейтинг каждого игрока, командная сумма, макро коэффициент и индекс доминации
/formula - Обоснование честности формулы
/top - Топ игроков текущего турнира
/hand - Ручное заполнения данных для использования формулы

Внимание: данные берутся из OpenDota API. Если матч закончился только что, подожди 1-2 минуты, пока сервер обработает информацию, либо зайди на страницу игры на сервисе opendota и нажми кнопку parse`
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
				bot.Send(msg)
			case "formula":
				msgText := `В формулах существует защита от “накрутки” внутриигровых параметров. Например, команда может решить играть на данный коэффициент и не заканчивать игру достаточно долго. В таком случае в формуле общая ценность игроков находится под логарифмом, а нанесенный урон (в случае с драфтом героев, которые наносят большой количество урона, не приводящего к убийству) - под гиперболическим тангенсом, соответственно чем больше эти параметры, тем меньше прирост для итогового коэффициента. То же самое можно сказать и про другие подобные метрики - перевес и опыт в минуту. Также в формуле учитывается другая проблема - потенциально команда может победить только за счет одного более сильного игрока. В таком случае команда получит меньшее значение командной суммы, чем команда с более равными показателями игроков. 

Все константы были получены исходя из предварительного анализа некоторого количества матчей и субъективной оценки важности каждого показателя.`
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
				bot.Send(msg)
			case "analyze", "details":
				args := update.Message.CommandArguments()
				if args == "" {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: введи ID матча после команды. Пример: /analyze 12345678"))
					return
				}

				id, err := strconv.ParseInt(args, 10, 64)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: ID матча должен быть числом."))
					return
				}

				match, err := opendota.GetMatch(int64(id))
				if err != nil {
					fmt.Print(err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Матч не найден, либо не запаршен на сервисе OpenDota"))
					bot.Send(msg)
					return
				}

				if len(match.Players) != 10 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Матч не запаршен на сервисе OpenDota, либо другая ошибка со стороны сервиса. проверьте парс и попробуйте еще раз"))
					bot.Send(msg)
					return
				}
				if len(match.Players[0].GoldTimes) == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Матч не запаршен на сервисе OpenDota, либо другая ошибка со стороны сервиса. проверьте парс и попробуйте еще раз"))
					bot.Send(msg)
					return
				}
				if len(match.GoldAdv) == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Матч не запаршен на сервисе OpenDota, либо другая ошибка со стороны сервиса. проверьте парс и попробуйте еще раз"))
					bot.Send(msg)
					return
				}
				if !match.IsRadWin {
					winnerText = "🏆*Победитель*: Dire\n"
				}
				if update.Message.Command() == "analyze" {
					FaR, _, _, _, _ := fal.FaCoef(match, match.Players[:5], match.TowersD, 1)
					FaD, _, _, _, _ := fal.FaCoef(match, match.Players[5:], match.TowersR, -1)

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(`%s
Коэффициент ANTON для Radiant - %f\n Коэффициент ANTON для Dire - %f`, winnerText, FaR, FaD))
					msg.ParseMode = "Markdown"
					bot.Send(msg)

				} else {
					instruments.PlayersHeroesIdToNames(match.Players)
					FaR, tSumR, macroR, domR, playersR := fal.FaCoef(match, match.Players[:5], match.TowersD, 1)
					FaD, tSumD, macroD, domD, playersD := fal.FaCoef(match, match.Players[5:], match.TowersR, -1)

					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf(`%s
*Коэффициент ANTON для Radiant* - %.4f,
*Коэффициент ANTON для Dire* - %.4f

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
%d. *%s* - %.4f`, winnerText, FaR, FaD, tSumR, tSumD, macroR, macroD, domR, domD, 1, match.Players[0].HeroName, playersR[0], 2, match.Players[1].HeroName, playersR[1],
						3, match.Players[2].HeroName, playersR[2], 4, match.Players[3].HeroName, playersR[3], 5, match.Players[4].HeroName, playersR[4],
						1, match.Players[5].HeroName, playersD[0], 2, match.Players[6].HeroName, playersD[1], 3, match.Players[7].HeroName, playersD[2],
						4, match.Players[8].HeroName, playersD[3], 5, match.Players[9].HeroName, playersD[4]))
					msg.ParseMode = "Markdown"
					bot.Send(msg)
				}
			case "ladder":
				if !instruments.IsAdmin(update.Message.From.ID) {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав администратора")
					bot.Send(msg)
					return
				}
				args := update.Message.CommandArguments()
				if args == "" {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: введи ID матча после команды. Пример: /ladder 12345678"))
					return
				}
				allPlayers, err := ladder.LoadLadder()
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при загрузке файла со всеми игроками. Напиши чепухе")
					bot.Send(msg)
					return
				}
				id, err := strconv.ParseInt(args, 10, 64)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка: ID матча должен быть числом."))
					return
				}
				match, err := opendota.GetMatch(int64(id))
				if err != nil {
					fmt.Print(err)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Матч не найден, либо не запаршен на сервисе OpenDota"))
					bot.Send(msg)
					return
				}

				if len(match.Players) != 10 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Матч не запаршен на сервисе OpenDota, либо другая ошибка со стороны сервиса. проверьте парс и попробуйте еще раз"))
					bot.Send(msg)
					return
				}
				if len(match.Players[0].GoldTimes) == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Матч не запаршен на сервисе OpenDota, либо другая ошибка со стороны сервиса. проверьте парс и попробуйте еще раз"))
					bot.Send(msg)
					return
				}
				if len(match.GoldAdv) == 0 {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Матч не запаршен на сервисе OpenDota, либо другая ошибка со стороны сервиса. проверьте парс и попробуйте еще раз"))
					bot.Send(msg)
					return
				}
				if ladder.IsMatchProcessed(id) {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("⚠️ Матч уже загружен в систему")))
					return
				}
				_, _, _, _, playersR := fal.FaCoef(match, match.Players[:5], match.TowersD, 1)
				_, _, _, _, playersD := fal.FaCoef(match, match.Players[5:], match.TowersR, -1)

				pool := ladder.CalculateMatchPTS(allPlayers, match.Players[:5], match.Players[5:], match.IsRadWin)
				if !match.IsRadWin {
					pool = -pool
				}
				allPlayers = ladder.ApplyMatchResults(allPlayers, match.Players[:5], pool, instruments.SumFa(playersR), playersR)
				allPlayers = ladder.ApplyMatchResults(allPlayers, match.Players[5:], -pool, instruments.SumFa(playersD), playersD)
				err = ladder.SaveLadder(allPlayers)
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("⚠️ Ошибка при загрузке матча. Попробуйте еще")))
					return
				}
				ladder.MarkMatchProcessed(id)
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("✅ Матч успешно загружен в систему. Итоговое изменение птс команды Radiant - %d", pool)))
			case "top":
				allPlayers, err := ladder.LoadLadder()
				if err != nil {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при загрузке файла со всеми игроками.")
					bot.Send(msg)
					return
				}
				msgText := ladder.GetTopFormatted(allPlayers)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
				msg.ParseMode = "markdown"
				bot.Send(msg)
			case "hand":
				setAwaiting(update.Message.Chat.ID, true)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID,
					"📋 Пришли JSON матча следующим сообщением — либо вставь его текстом, либо отправь .json файл.\n\n"+
						"Шаблон JSON (заполни для всех 10 игроков):\n"+
						"```\n"+
						"{\n"+
						"  \"duration\": 0,\n"+
						"  \"radiant_win\": true,\n"+
						"  \"tower_status_radiant\": 0,\n"+
						"  \"tower_status_dire\": 0,\n"+
						"  \"radiant_gold_adv\": [0, 0, ...],\n"+
						"  \"players\": [\n"+
						"    {\n"+
						"      \"hero_id\": 0,\n"+
						"      \"kills\": 0,\n"+
						"      \"deaths\": 0,\n"+
						"      \"assists\": 0,\n"+
						"      \"last_hits\": 0,\n"+
						"      \"hero_damage\": 0,\n"+
						"      \"xp_per_min\": 0,\n"+
						"      \"gold_t\": [0, 0, ...]\n"+
						"    }\n"+
						"  ]\n"+
						"}\n"+
						"```\n"+
						"_players[0-4] — Radiant, players[5-9] — Dire_\n\n"+
						"Чтобы отменить, отправь cancel.")
				msg.ParseMode = "Markdown"
				bot.Send(msg)
			default:
				msgText := "неизвестная команда"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
				bot.Send(msg)
			}
		}
	}
}
