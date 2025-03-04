package main

import (
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SentMessage struct {
	MessageID int
	ChatID    int64
}

type MessageManager struct {
	data       map[int]SentMessage
	deleteChan chan int
	mu         sync.Mutex
}

func NewMessageManager(bot *tgbotapi.BotAPI) *MessageManager {
	messageManager := &MessageManager{
		data:       make(map[int]SentMessage),
		deleteChan: make(chan int),
	}
	go messageManager.clearExpiredMessage(bot)
	return messageManager
}

func (mm *MessageManager) clearExpiredMessage(bot *tgbotapi.BotAPI) {
	for k := range mm.deleteChan {
		msg, _ := mm.get(k)
		msgDeleteConf := tgbotapi.NewDeleteMessage(msg.ChatID, msg.MessageID)
		bot.Request(msgDeleteConf)
		mm.mu.Lock()
		delete(mm.data, k)
		mm.mu.Unlock()

	}
}

func (mm *MessageManager) Add(key int, value SentMessage) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.data[key] = value
	time.AfterFunc(1*time.Minute,
		func() {
			mm.deleteChan <- key
		})
}

func (lp *MessageManager) get(key int) (SentMessage, bool) {
	lp.mu.Lock()
	defer lp.mu.Unlock()
	val, ok := lp.data[key]
	return val, ok
}
