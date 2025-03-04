package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/RemyJohnny/timetable/mdb"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	/* err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	} */
	logFile, err := os.OpenFile("timetable.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)

	clientOptions := options.Client().ApplyURI(os.Getenv("TIMETABLE_MONGODB_STRING"))
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	db := mdb.Db{
		LectureCollection: client.Database("timetable").Collection("lecture"),
	}

	// Initialize the bot with your token
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TIMETABLE_TG_BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	admins := strings.Split(os.Getenv("TIMETABLE_ADMINS_USERID"), "|")

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Set up an update config to listen for new messages
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	var lectureInput = make(map[int64]mdb.Lecture)
	var lectureUpdate = make(map[int64]UpdateLecture)
	var lectureDelete = make(map[int64]string)

	var mm = NewMessageManager(bot)

	for update := range updates {
		if update.Message == nil { // ignore non-messages
			continue
		}

		// Respond to commands

		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		text := strings.ToLower(update.Message.Text)
		command := update.Message.Command()
		switch command {
		case "addlecture":
			if Auth(admins, strconv.FormatInt(userID, 10)) {
				lectureInput[userID] = mdb.Lecture{}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "select the week for the lecture ( 0 for all)")
				msg.ReplyMarkup = GenMenu(mdb.Weeks, false)
				bot.Send(msg)
			}
		case "editlecture":
			if Auth(admins, strconv.FormatInt(userID, 10)) {
				lectureUpdate[userID] = UpdateLecture{}
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Enter the ID of the lecture you want to edit: ")
				bot.Send(msg)
			}
		case "deletelecture":
			if Auth(admins, strconv.FormatInt(userID, 10)) {
				lectureDelete[userID] = ""
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Enter the ID of the lecture you want to delete: ")
				bot.Send(msg)
			}
		case "today":
			argStr, _ := strings.CutPrefix(strings.TrimSpace(text), "/today")
			arg := ParseArgs(argStr)
			sendToday(&db, chatID, bot, arg, false, mm)
		case "tomorrow":
			argStr, _ := strings.CutPrefix(strings.TrimSpace(text), "/tomorrow")
			arg := ParseArgs(argStr)
			sendToday(&db, chatID, bot, arg, true, mm)
		case "thisweek":
			argStr, _ := strings.CutPrefix(strings.TrimSpace(text), "/thisweek")
			arg := ParseArgs(argStr)
			SendWeek(&db, chatID, bot, arg, false, mm)
		case "nextweek":
			argStr, _ := strings.CutPrefix(strings.TrimSpace(text), "/nextweek")
			arg := ParseArgs(argStr)
			SendWeek(&db, chatID, bot, arg, true, mm)
			//russian cmd
		case "сегодня":
			argStr, _ := strings.CutPrefix(strings.TrimSpace(text), "/сегодня")
			arg := ParseArgs(argStr)
			sendToday(&db, chatID, bot, arg, false, mm)
		case "завтра":
			argStr, _ := strings.CutPrefix(strings.TrimSpace(text), "/завтра")
			arg := ParseArgs(argStr)
			sendToday(&db, chatID, bot, arg, true, mm)
		case "текущая-неделя":
			argStr, _ := strings.CutPrefix(strings.TrimSpace(text), "/текущая-неделя")
			arg := ParseArgs(argStr)
			SendWeek(&db, chatID, bot, arg, false, mm)
		case "следующая-неделя":
			argStr, _ := strings.CutPrefix(strings.TrimSpace(text), "/следующая-неделя")
			arg := ParseArgs(argStr)
			SendWeek(&db, chatID, bot, arg, true, mm)
		case "help":
			helpTxt := "*/today | /сегодня* `команда возвращает расписание на сегодня`\n\n*/tomorrow | /завтра* `команда возвращает расписание на завтра`\n\n*/thisweek | /текущая-неделя* `команда возвращает расписание на текущую неделю`\n\n*/nextweek | /следующая-неделя* `команда возвращает расписание на следующую неделю`\n\n\n"
			flagTxt := "*-l*  : `Отображает полное имя предмета и имя преподавателя. имя предмета по умолчанию сокращается`\n\n*-1*  : `возвращает расписание для подгруппы 1 . по умолчанию` \n\n*-2*  : `возвращает расписание для подгруппы 2 `\n\n*-all*  : `возвращает расписание для всей подгруппы`\n\n"
			expTxt := "*-Пример-*\n    /сегодня -l -2\n`Возвращает расписание на сегодня и для подгруппы 2 с именем лектора и полным именем предмета.`"
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpTxt+"`Добавьте флаги в команду, чтобы изменить, как и что возвращается. флаги:`\n"+flagTxt+expTxt)
			msg.ParseMode = tgbotapi.ModeMarkdown
			_, err := bot.Send(msg)
			if err != nil {
				log.Printf("error: %v", err)
			}
		default:
			HandleLectureInput(&db, lectureInput, &update, bot)
			HandleLectureUpdate(&db, lectureUpdate, &update, bot)
			HandleLectureDelete(&db, lectureDelete, &update, bot)
		}
	}
}
