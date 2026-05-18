package main

import (
	"coefbot/bothandler"
	"coefbot/instruments"
	"sync"

	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	instruments.InitHeroes()
	token := os.Getenv("BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Panic(err)
	}

	go func() {
		<-stop
		fmt.Println("\nПолучен сигнал остановки. Завершаем работу...")
		cancel()
	}()

	for {
		select {
		case <-ctx.Done():
			goto shutdown
		case update, ok := <-updates:
			if !ok {
				goto shutdown
			}
			wg.Add(1)
			go func(u tgbotapi.Update) {
				defer wg.Done()
				bothandler.MessageHandler(bot, u)
			}(update)
		}
	}

shutdown:
	wg.Wait()
	fmt.Println("Все процессы завершены. Выход.")
}
