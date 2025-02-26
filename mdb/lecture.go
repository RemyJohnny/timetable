package mdb

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// inserts new lecture to the database
func (d *Db) InsertLecture(lecture Lecture) error {
	week, _ := strconv.Atoi(lecture.Week)
	if week < 0 || week > 4 {
		return fmt.Errorf("error: week must be between 0 - 4")
	}
	result, err := d.LectureCollection.InsertOne(context.TODO(), lecture)
	if err != nil {
		return fmt.Errorf("error inserting lecture: %w", err)
	}
	log.Printf("Inserted lecture with ID: %v\n", result.InsertedID)
	return nil
}

func (d *Db) UpdateLecture(ID primitive.ObjectID, lecture Lecture) error {
	result, err := d.LectureCollection.UpdateByID(context.TODO(), ID, bson.M{"$set": lecture})
	if err != nil {
		return fmt.Errorf("error updating lecture: %w", err)
	}
	if result.MatchedCount < 1 {
		return fmt.Errorf("lecture not found")
	}
	log.Printf("updated: %v lecture\n", result.MatchedCount)
	return nil
}

func (d *Db) GetLecture(lectureID string) (Lecture, error) {
	ID, err := primitive.ObjectIDFromHex(lectureID)
	if err != nil {
		return Lecture{}, fmt.Errorf("error converting ObjectID from Hex: %w", err)
	}
	var lecture Lecture
	filter := bson.M{"_id": ID}
	err = d.LectureCollection.FindOne(context.TODO(), filter).Decode(&lecture)
	if err != nil {
		return Lecture{}, fmt.Errorf("error: %w", err)
	}

	return lecture, nil
}

func (d *Db) GetLectures(filter bson.M) ([]Lecture, error) {
	cursor, err := d.LectureCollection.Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("error getting lectures: %w", err)
	}
	defer cursor.Close(context.TODO())

	var lectures []Lecture
	if err = cursor.All(context.TODO(), &lectures); err != nil {
		return nil, fmt.Errorf("error decoding lecture: %w", err)
	}

	slices.SortStableFunc(lectures, func(a, b Lecture) int {
		if a.Week < b.Week {
			return -1
		} else if a.Week > b.Week {
			return 1
		}
		return 0
	})

	slices.SortStableFunc(lectures, func(a, b Lecture) int {
		if a.Time < b.Time {
			return -1
		} else if a.Time > b.Time {
			return 1
		}
		return 0
	})

	return lectures, nil
}

func (d *Db) DeleteLecture(lectureID string) error {
	ID, err := primitive.ObjectIDFromHex(lectureID)
	if err != nil {
		return fmt.Errorf("error converting ObjectID from Hex: %w", err)
	}
	deleteFilter := bson.M{"_id": ID}
	result, err := d.LectureCollection.DeleteOne(context.TODO(), deleteFilter)
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("lecture [ %s ] not found", lectureID)
	}
	return nil
}
