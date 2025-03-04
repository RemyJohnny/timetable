package mdb

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Lecture struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Week     string             `bson:"week"`
	Subject  string             `bson:"subject"`
	Time     int                `bson:"time"`
	Type     string             `bson:"type"`
	Day      int                `bson:"day"`
	Room     string             `bson:"room"`
	Lecturer string             `bson:"lecturer"`
	SubGroup string             `bson:"sub_group"`
}

type Db struct {
	LectureCollection *mongo.Collection
}

type Subject struct {
	Name     string
	Key      string
	Lecturer string
}
type Period struct {
	start string
	end   string
}

func (p Period) String() string {
	return fmt.Sprintf("%s-%s", p.start, p.end)
}

var Subjects = map[string]Subject{
	"ТЭ":    {Name: "Техническая Электроника", Key: "ТЭ", Lecturer: "Половеня С.И"},
	"ОИкТ":  {Name: "Основы Инфокоммуникационных Технологий", Key: "ОИкТ", Lecturer: "Дулькевич А.И"},
	"ФК":    {Name: "Физическая Культура", Key: "ФК", Lecturer: "Байко О.М"},
	"ТПИкС": {Name: "Технологии Программирования Инфокоммуникационных Систем", Key: "ТПИкС", Lecturer: "Рябычина О.П"},
	"ТЭЦ":   {Name: "Теория Электрических Цепей", Key: "ТЭЦ", Lecturer: "Кочергина О.В"},
	"АЯ":    {Name: "Английский Язык", Key: "АЯ", Lecturer: "Мышелова Н.И"},
	"ПУвИО": {Name: "Психология Управления в Информационном Обществе", Key: "ПУвИО", Lecturer: "Христина Л.Ф"},
	"ОТФ":   {Name: "Основы Теории Фильтрации", Key: "ОТФ", Lecturer: "Киевец Н.Г"},
	"БЯ":    {Name: "Белорусский Язык", Key: "БЯ", Lecturer: "Чуприна Е.А"},
	"ОЦС":   {Name: "Основы Цифровой Схемотехники", Key: "ОЦС", Lecturer: "Постельняк  А.А"},
	"ОМО":   {Name: "Основы Машинного Обучения", Key: "ОМО", Lecturer: "Колодный В.Б"},
}

var Periods = map[int]Period{
	1: {start: "8:00", end: "9:40"},
	2: {start: "9:55", end: "11:35"},
	3: {start: "12:15", end: "13:55"},
	4: {start: "14:10", end: "15:50"},
	5: {start: "16:20", end: "18:00"},
	6: {start: "18:15", end: "19:55"},
}

var Days = map[int]string{
	1: "Понедельник",
	2: "Вторник",
	3: "Среда",
	4: "Четверг",
	5: "Пятница",
	6: "Суббота",
}
var Types = map[string]int{
	"ЛР": 1,
	"ПЗ": 2,
	"ЛК": 4,
	"-":  5,
}
var Weeks = map[string]int{
	"0": 1,
	"1": 2,
	"2": 3,
	"3": 4,
	"4": 5,
}
var SubGroup = map[string]int{
	"0": 1,
	"1": 2,
	"2": 3,
}

type Args struct {
	Long  bool
	Group string
}

var CmdOpts = map[string]int{
	"-l":   1,
	"-1":   2,
	"-2":   3,
	"-all": 4,
}
