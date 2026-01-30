package main

import (
	"coefbot/bothandler"
	"coefbot/instruments"
	"sync"

	"log"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"context"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	instruments.InitHeroes()
	bot, err := tgbotapi.NewBotAPI("YOUR_BOT_API")
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
		case update, ok := <- updates:
			if !ok {goto shutdown}
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