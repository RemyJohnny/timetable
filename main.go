package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/RemyJohnny/timetable/mdb"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

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

	for update := range updates {
		if update.Message == nil { // ignore non-messages
			continue
		}

		// Respond to commands

		userID := update.Message.From.ID
		text := strings.ToLower(update.Message.Text)
		switch update.Message.Command() {
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
			sendToday(&db, userID, bot, arg, false)
		case "tomorrow":
			argStr, _ := strings.CutPrefix(strings.TrimSpace(text), "/tomorrow")
			arg := ParseArgs(argStr)
			sendToday(&db, userID, bot, arg, true)
		case "thisweek":
			argStr, _ := strings.CutPrefix(strings.TrimSpace(text), "/week")
			arg := ParseArgs(argStr)
			SendWeek(&db, userID, bot, arg, false)
		case "nextweek":
			argStr, _ := strings.CutPrefix(strings.TrimSpace(text), "/nextweek")
			arg := ParseArgs(argStr)
			SendWeek(&db, userID, bot, arg, true)
		case "help":
			helpTxt := "*/today* `command gets the timetable for today`\n\n*/tomorrow* `command gets the timetable for tomorrow`\n\n*/thisweek* `command gets the timetable for the current week`\n\n*/nextweek* `command gets the timetable for next week`\n\n\n"
			flagTxt := "*-l*  : `Displays subjects full Name. subjects name are abbrivated by default`\n\n*-1*  : `Timetable for subgroup 1 are returned. default` \n\n*-2*  : `Timetable for subgroup 2 are returned`\n\n"
			expTxt := "*-Example-*\n    /today -l -2\n`Returns the timetable for today and for subgroup 2 with Lecturer's Name and full Subject Name`"
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, helpTxt+"`Add flags to command modify how and what is being returned. flags:`\n"+flagTxt+expTxt)
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
