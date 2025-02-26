package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RemyJohnny/timetable/mdb"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// determines the current academic week
// semesterStartDate (YYYY-MM-DD format)
func GetCurrentWeek(semesterStartDate string) (int, error) {
	startDate, err := time.Parse("2006-01-02", semesterStartDate)
	if err != nil {
		return 0, fmt.Errorf("error parsing date: %w", err)
	}

	currentDate := time.Now()

	duration := currentDate.Sub(startDate)

	weeksSinceStart := int(duration.Hours() / (24 * 7))

	currentAcademicWeek := (weeksSinceStart % 4) + 1

	if currentAcademicWeek < 1 {
		return 0, fmt.Errorf("semester has not started")
	}

	return currentAcademicWeek, nil

}

func ParseArgs(str string) mdb.Args {
	var arg mdb.Args
	arrStr := strings.Split(str, " ")
	for _, cmd := range arrStr {
		key := strings.TrimSpace(cmd)
		if _, ok := mdb.CmdOpts[key]; ok {
			switch key {
			case "-l":
				arg.Long = true
			case "-1":
				arg.Group = "1"
			case "-2":
				arg.Group = "2"
			default:
				continue
			}
		}
	}
	return arg
}

type UpdateLecture struct {
	OldLecture mdb.Lecture
	NewLecture mdb.Lecture
}

type LectureInput map[int64]mdb.Lecture
type LectureUpdate map[int64]UpdateLecture
type LectureDelete map[int64]string

func genSubjectMenu(subjects map[string]mdb.Subject, skipkey bool) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for _, subject := range subjects {
		row := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(subject.Name),
		)
		rows = append(rows, row)
	}
	if skipkey {
		skipkey := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("skip"),
		)

		rows = append(rows, skipkey)
	}
	return tgbotapi.NewOneTimeReplyKeyboard(rows...)
}
func GenMenu(menu map[string]int, skipkey bool) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for key := range menu {
		row := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(key),
		)
		rows = append(rows, row)
	}
	if skipkey {
		skipkey := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("skip"),
		)

		rows = append(rows, skipkey)
	}
	return tgbotapi.NewOneTimeReplyKeyboard(rows...)
}
func genPeriodMenu(periods map[int]mdb.Period, skipkey bool) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for _, period := range periods {
		row := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(period.String()),
		)
		rows = append(rows, row)
	}
	if skipkey {
		skipkey := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("skip"),
		)

		rows = append(rows, skipkey)
	}
	return tgbotapi.NewOneTimeReplyKeyboard(rows...)
}
func GenDaysMenu(days map[int]string, skipkey bool) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for _, val := range days {
		row := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(val),
		)
		rows = append(rows, row)
	}
	if skipkey {
		skipkey := tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("skip"),
		)

		rows = append(rows, skipkey)
	}
	return tgbotapi.NewOneTimeReplyKeyboard(rows...)
}

func HandleLectureInput(db *mdb.Db, lectureInput LectureInput, update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	if lecture, exists := lectureInput[userID]; exists {
		if text == "cancel" {
			delete(lectureInput, userID)
			msg := tgbotapi.NewMessage(chatID, "Lecture insert cancelled")
			bot.Send(msg)
			return
		}
		if lecture.Week == "" {
			if _, ok := mdb.Weeks[text]; ok {
				lecture.Week = text
				lectureInput[userID] = lecture
				msg := tgbotapi.NewMessage(chatID, "Great! Now, choose the subject")
				msg.ReplyMarkup = genSubjectMenu(mdb.Subjects, false)
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid option please select the week for the lecture ( 0 for all)")
				msg.ReplyMarkup = GenMenu(mdb.Weeks, false)
				bot.Send(msg)
			}
		} else if lecture.Subject == "" {
			valid := false
			for key, subject := range mdb.Subjects {
				if text == subject.Name {
					lecture.Subject = subject.Key
					lecture.Lecturer = subject.Lecturer
					lectureInput[userID] = lecture
					valid = true
					msg := tgbotapi.NewMessage(chatID, "select the type of the lecture")
					msg.ReplyMarkup = GenMenu(mdb.Types, false)
					bot.Send(msg)
					fmt.Println(key)
					break
				}
			}
			if !valid {
				msg := tgbotapi.NewMessage(chatID, "invalid subject please choose a valid subject")
				msg.ReplyMarkup = genSubjectMenu(mdb.Subjects, false)
				bot.Send(msg)
			}
		} else if lecture.Type == "" {
			if _, ok := mdb.Types[text]; ok {
				lecture.Type = text
				lectureInput[userID] = lecture
				msg := tgbotapi.NewMessage(chatID, "select the day of the week for the lecture")
				msg.ReplyMarkup = GenDaysMenu(mdb.Days, false)
				bot.Send(msg)
			} else {
				msg := tgbotapi.NewMessage(chatID, "Invalid please select the type of the lecture")
				msg.ReplyMarkup = GenMenu(mdb.Types, false)
				bot.Send(msg)
			}
		} else if lecture.Day == 0 {
			valid := false
			for key, day := range mdb.Days {
				if text == day {
					lecture.Day = key
					lectureInput[userID] = lecture
					valid = true
					msg := tgbotapi.NewMessage(chatID, "Enter the room for the lecture")
					bot.Send(msg)
					break
				}
			}
			if !valid {
				msg := tgbotapi.NewMessage(chatID, "Invalid please select the day of the week for the lecture")
				msg.ReplyMarkup = GenDaysMenu(mdb.Days, false)
				bot.Send(msg)
			}
		} else if lecture.Room == "" {
			lecture.Room = text
			lectureInput[userID] = lecture
			msg := tgbotapi.NewMessage(chatID, "select the period of the lecture")
			msg.ReplyMarkup = genPeriodMenu(mdb.Periods, false)
			bot.Send(msg)
		} else if lecture.Time == 0 {
			valid := false
			for key, period := range mdb.Periods {
				if text == period.String() {
					lecture.Time = key
					lectureInput[userID] = lecture
					valid = true
					msg := tgbotapi.NewMessage(chatID, "select the subGroup to take the lecture ( 0 for all )")
					msg.ReplyMarkup = GenMenu(mdb.SubGroup, false)
					bot.Send(msg)
					break
				}
			}
			if !valid {
				msg := tgbotapi.NewMessage(chatID, "Invalid period select the period of the lecture")
				msg.ReplyMarkup = genPeriodMenu(mdb.Periods, false)
				bot.Send(msg)
			}
		} else if lecture.SubGroup == "" {
			if _, ok := mdb.SubGroup[text]; ok {
				lecture.SubGroup = text
				lectureInput[userID] = lecture
				err := db.InsertLecture(lecture)
				if err != nil {
					msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("error: %v", err))
					bot.Send(msg)
					delete(lectureInput, userID)
				}
				log.Printf("New lecture : %+v", lecture)
				msg := tgbotapi.NewMessage(chatID, "Added successfully")
				bot.Send(msg)
				delete(lectureInput, userID)
			} else {
				msg := tgbotapi.NewMessage(chatID, "Invalid option please select the subGroup to take the lecture ( 0 for all )")
				msg.ReplyMarkup = GenMenu(mdb.SubGroup, false)
				bot.Send(msg)
			}

		}
	}
}

func HandleLectureUpdate(db *mdb.Db, lectureUpdate LectureUpdate, update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	text := strings.ToLower(strings.TrimSpace(update.Message.Text))

	if edit, exists := lectureUpdate[userID]; exists {
		if text == "cancel" {
			delete(lectureUpdate, userID)
			msg := tgbotapi.NewMessage(chatID, "Lecture update cancelled")
			bot.Send(msg)
		}

		if edit.OldLecture.ID.IsZero() {
			_, err := primitive.ObjectIDFromHex(text)
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, "invalid lectureID")
				bot.Send(msg)
			} else {
				l, err := db.GetLecture(text)
				if err != nil {
					log.Println(err)
					msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("error : %v", err))
					bot.Send(msg)
				} else {
					edit.OldLecture = l
					lectureUpdate[userID] = edit
					msg := tgbotapi.NewMessage(chatID, "Please select the lecture week \nReply skip to use old week:")
					msg.ReplyMarkup = genSubjectMenu(mdb.Subjects, true)
					bot.Send(msg)
				}
			}
		} else if edit.NewLecture.Week == "" {
			if text == "skip" {
				edit.NewLecture.Week = edit.OldLecture.Week
				lectureUpdate[userID] = edit
				msg := tgbotapi.NewMessage(chatID, "Great! Now, select the new subject name  \nReply skip to use old subject name")
				msg.ReplyMarkup = genSubjectMenu(mdb.Subjects, true)
				bot.Send(msg)
			} else {
				if _, ok := mdb.Weeks[text]; ok {
					edit.NewLecture.Week = text
					lectureUpdate[userID] = edit
					msg := tgbotapi.NewMessage(chatID, "Great! Now, choose the subject")
					msg.ReplyMarkup = genSubjectMenu(mdb.Subjects, true)
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Invalid option please select the week for the lecture ( 0 for all)")
					msg.ReplyMarkup = GenMenu(mdb.Weeks, true)
					bot.Send(msg)
				}
			}
		} else if edit.NewLecture.Subject == "" {
			if text == "skip" {
				edit.NewLecture.Subject = edit.OldLecture.Subject
				lectureUpdate[userID] = edit
				msg := tgbotapi.NewMessage(chatID, "select the new type of the lecture \nReply skip to use the old type")
				msg.ReplyMarkup = GenMenu(mdb.Types, true)
				bot.Send(msg)
			} else {
				valid := false
				for key, subject := range mdb.Subjects {
					if text == subject.Name {
						edit.NewLecture.Subject = subject.Key
						edit.NewLecture.Lecturer = subject.Lecturer
						lectureUpdate[userID] = edit
						valid = true
						msg := tgbotapi.NewMessage(chatID, "select the type of the lecture")
						msg.ReplyMarkup = GenMenu(mdb.Types, true)
						bot.Send(msg)
						fmt.Println(key)
						break
					}
				}
				if !valid {
					msg := tgbotapi.NewMessage(chatID, "invalid subject please choose a valid subject")
					msg.ReplyMarkup = genSubjectMenu(mdb.Subjects, true)
					bot.Send(msg)
				}
			}
		} else if edit.NewLecture.Type == "" {
			if text == "skip" {
				edit.NewLecture.Type = edit.OldLecture.Type
				lectureUpdate[userID] = edit
				msg := tgbotapi.NewMessage(chatID, "select the new Day of the week for the lecture \nReply skip to use the old lecture Day")
				msg.ReplyMarkup = GenDaysMenu(mdb.Days, true)
				bot.Send(msg)
			} else {
				if _, ok := mdb.Types[text]; ok {
					edit.NewLecture.Type = text
					lectureUpdate[userID] = edit
					msg := tgbotapi.NewMessage(chatID, "select the new Day of the week for the lecture \nReply skip to use the old lecture Day")
					msg.ReplyMarkup = GenDaysMenu(mdb.Days, true)
					bot.Send(msg)
				} else {
					msg := tgbotapi.NewMessage(chatID, "Invalid please select the type of the lecture")
					msg.ReplyMarkup = GenMenu(mdb.Types, true)
					bot.Send(msg)
				}
			}
		} else if edit.NewLecture.Day == 0 {
			if text == "skip" {
				edit.NewLecture.Day = edit.OldLecture.Day
				lectureUpdate[userID] = edit
				msg := tgbotapi.NewMessage(chatID, "Enter the new Auditorium for the lecture \nReply skip to use the old lecture Auditorium")
				msg.ReplyMarkup = tgbotapi.NewOneTimeReplyKeyboard(tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton("skip")))
				bot.Send(msg)
			} else {
				valid := false
				for key, day := range mdb.Days {
					if text == day {
						edit.NewLecture.Day = key
						lectureUpdate[userID] = edit
						valid = true
						msg := tgbotapi.NewMessage(chatID, "Enter the room for the lecture")
						bot.Send(msg)
						break
					}
				}
				if !valid {
					msg := tgbotapi.NewMessage(chatID, "Invalid please select the day of the week for the lecture")
					msg.ReplyMarkup = GenDaysMenu(mdb.Days, true)
					bot.Send(msg)
				}
			}
		} else if edit.NewLecture.Room == "" {
			if text == "skip" {
				edit.NewLecture.Room = edit.OldLecture.Room
				lectureUpdate[userID] = edit
				msg := tgbotapi.NewMessage(chatID, "select the new period for the lecture \nReply skip to use the old period")
				msg.ReplyMarkup = genPeriodMenu(mdb.Periods, true)
				bot.Send(msg)
			} else {
				edit.NewLecture.Room = text
				lectureUpdate[userID] = edit
				msg := tgbotapi.NewMessage(chatID, "select the new period for the lecture")
				msg.ReplyMarkup = genPeriodMenu(mdb.Periods, true)
				bot.Send(msg)
			}
		} else if edit.NewLecture.Time == 0 {
			if text == "skip" {
				edit.NewLecture.Time = edit.OldLecture.Time
				lectureUpdate[userID] = edit
				msg := tgbotapi.NewMessage(chatID, "select the new SubGroup to take the lecture (0 for all ) \nReply skip to use the old subGroup")
				msg.ReplyMarkup = GenMenu(mdb.SubGroup, true)
				bot.Send(msg)
			} else {
				valid := false
				for key, period := range mdb.Periods {
					if text == period.String() {
						edit.NewLecture.Time = key
						lectureUpdate[userID] = edit
						valid = true
						msg := tgbotapi.NewMessage(chatID, "select the new subGroup to take the lecture ( 0 for all ) \nReply skip to use the old subGroup")
						msg.ReplyMarkup = GenMenu(mdb.SubGroup, true)
						bot.Send(msg)
						break
					}
				}
				if !valid {
					msg := tgbotapi.NewMessage(chatID, "Invalid period select the period of the lecture")
					msg.ReplyMarkup = genPeriodMenu(mdb.Periods, true)
					bot.Send(msg)
				}
			}
		} else if edit.NewLecture.SubGroup == "" {
			if text == "skip" {
				edit.NewLecture.SubGroup = edit.OldLecture.SubGroup
				lectureUpdate[userID] = edit
				err := db.UpdateLecture(edit.OldLecture.ID, edit.NewLecture)
				if err != nil {
					msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("error: %v", err))
					bot.Send(msg)
					delete(lectureUpdate, userID)
				}
				log.Printf("updated lecture : %v", edit.OldLecture.ID.Hex())
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("updated [ %v ] successfully", edit.OldLecture.ID))
				bot.Send(msg)
				delete(lectureUpdate, userID)
			} else {
				if _, ok := mdb.SubGroup[text]; ok {
					edit.NewLecture.SubGroup = text
					lectureUpdate[userID] = edit
					err := db.UpdateLecture(edit.OldLecture.ID, edit.NewLecture)
					if err != nil {
						msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("error: %v", err))
						bot.Send(msg)
						delete(lectureUpdate, userID)
					}
					log.Printf("updated lecture : %v", edit.OldLecture.ID.Hex())
					msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("updated [ %v ] successfully", edit.OldLecture.ID))
					bot.Send(msg)
					delete(lectureUpdate, userID)
				} else {
					msg := tgbotapi.NewMessage(chatID, "Invalid option please select the subGroup to take the lecture ( 0 for all )")
					msg.ReplyMarkup = GenMenu(mdb.SubGroup, true)
					bot.Send(msg)
				}
			}
		}
	}
}

func HandleLectureDelete(db *mdb.Db, lectureDelete LectureDelete, update *tgbotapi.Update, bot *tgbotapi.BotAPI) {
	userID := update.Message.From.ID
	chatID := update.Message.Chat.ID
	text := strings.ToLower(strings.TrimSpace(update.Message.Text))

	if id, exists := lectureDelete[userID]; exists {
		if text == "cancel" {
			delete(lectureDelete, userID)
			msg := tgbotapi.NewMessage(chatID, "Lecture delete cancelled")
			bot.Send(msg)
		}
		if id == "" {
			id = text
			lectureDelete[userID] = id
			err := db.DeleteLecture(lectureDelete[userID])
			if err != nil {
				msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("error : %v", err))
				bot.Send(msg)
			}
			delete(lectureDelete, userID)
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Deleted lecture [ %v ] successfully", id))
			bot.Send(msg)
		}
	}
}

func Auth(admins []string, userID string) bool {
	for i := 0; i < len(admins); i++ {
		if admins[i] == userID {
			return true
		}
	}
	return false
}

func FormatLecture(lecture mdb.Lecture, opt mdb.Args) string {
	var subject string
	line := "----------------------------------------"
	if opt.Long {
		subject = mdb.Subjects[lecture.Subject].Name
		return fmt.Sprintf("%v\n`%v | %v | %v | %v | %v`\n%v\n", line, mdb.Periods[lecture.Time], subject, lecture.Type, lecture.Room, lecture.Lecturer, line)

	} else {
		subject = mdb.Subjects[lecture.Subject].Key
		return fmt.Sprintf("%v\n`%v | %v | %v | %v`\n%v\n", line, mdb.Periods[lecture.Time], subject, lecture.Type, lecture.Room, line)
	}
}

func SendLectures(lectures []mdb.Lecture, day string, chatID int64, bot *tgbotapi.BotAPI, opt mdb.Args) {
	header := fmt.Sprintf("*%v*\n", day)
	var content string
	for _, lecture := range lectures {
		content += FormatLecture(lecture, opt)
	}
	msg := tgbotapi.NewMessage(chatID, header+content)
	msg.ParseMode = tgbotapi.ModeMarkdown
	bot.Send(msg)
}

func sendToday(db *mdb.Db, chatID int64, bot *tgbotapi.BotAPI, opt mdb.Args, tommorrow bool) {
	print := "No Lectures Today 🎊"
	week, _ := GetCurrentWeek(os.Getenv("SEMESTER_START_DATE"))
	day := int(time.Now().Weekday())
	if tommorrow {
		day += 1
		if day == 1 {
			day = 1
			week += 1
			if week > 4 {
				week = 1
			}
		}
		print = "No Lectures Tomorrow 🎊"
	}
	filter := bson.M{
		"$or": []bson.M{
			{"week": fmt.Sprint(week)},
			{"week": "0"}},
		"day": day,
	}
	if opt.Group != "" {
		filter = bson.M{
			"$and": []bson.M{
				{"$or": []bson.M{
					{"week": fmt.Sprint(week)},
					{"week": "0"}}},
				{"$or": []bson.M{
					{"sub_group": opt.Group},
					{"sub_group": "0"}},
				}},
			"day": day,
		}
	}
	lectures, err := db.GetLectures(filter)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("error : %v", err))
		bot.Send(msg)
	}
	if len(lectures) > 0 {
		SendLectures(lectures, mdb.Days[day], chatID, bot, opt)
	} else {
		msg := tgbotapi.NewMessage(chatID, print)
		bot.Send(msg)
	}

}

func SendWeek(db *mdb.Db, chatID int64, bot *tgbotapi.BotAPI, opt mdb.Args, nextWeek bool) {
	week, _ := GetCurrentWeek(os.Getenv("SEMESTER_START_DATE"))
	if nextWeek {
		week += 1
		if week > 4 {
			week = 1
		}
	}
	fmt.Printf("week: %v", week)
	daysKeys := []int{1, 2, 3, 4, 5, 6}
	if opt.Group == "" {
		opt.Group = "1"
	}
	filter := bson.M{
		"$and": []bson.M{
			{"$or": []bson.M{
				{"week": fmt.Sprint(week)},
				{"week": "0"}}},
			{"$or": []bson.M{
				{"sub_group": opt.Group},
				{"sub_group": "0"}},
			}},
	}
	lectures, err := db.GetLectures(filter)
	if err != nil {
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("error : %v", err))
		bot.Send(msg)
	}
	for _, v := range daysKeys {
		var day []mdb.Lecture
		for _, lecture := range lectures {
			if lecture.Day == v {
				day = append(day, lecture)
			}
		}
		if len(day) > 0 {
			SendLectures(day, mdb.Days[v], chatID, bot, opt)
		} else {
			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(" %v свободен🎊", mdb.Days[v]))
			bot.Send(msg)
		}

	}
}
