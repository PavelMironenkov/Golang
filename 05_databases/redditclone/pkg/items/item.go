package items

import (
	//"gopkg.in/mgo.v2/bson"
	bson "go.mongodb.org/mongo-driver/bson/primitive"
)

type Author struct {
	Username string `json:"username" bson:"Username"`
	ID       string `json:"id" bson:"ID"`
}

type Comment struct {
	Created   string `json:"created" bson:"Created"`
	Author    Author `json:"author" bson:"Author"`
	Comment   string `json:"body" bson:"Comment"`
	CommentID string `json:"id" bson:"CommentID"`
}

type Vote struct {
	UserID string `json:"user" bson:"UserID"`
	Vote   int8   `json:"vote" bson:"Vote"`
}

type Item struct {
	Score            int64         `json:"score" bson:"Score"`
	Views            uint64        `json:"views" bson:"Views"`
	Type             string        `json:"type" bson:"Type"`
	Title            string        `json:"title" bson:"Title"`
	Author           Author        `json:"author" bson:"Author"`
	Category         string        `json:"category" bson:"Category"`
	Text             string        `json:"text,omitempty" bson:"Text,omitempty"`
	URL              string        `json:"url,omitempty" bson:"URL,omitempty"`
	Votes            []Vote        `json:"votes" bson:"Votes"`
	Comments         []Comment     `json:"comments" bson:"Comments"`
	Created          string        `json:"created" bson:"Created"`
	UpvotePercentage uint8         `json:"upvotePercentage" bson:"UpvotePercentage"`
	PostID           bson.ObjectID `json:"id" bson:"_id"`
}

type ItemsRepo interface {
	GetAll() ([]*Item, error)
	GetByID(id string) (*Item, error)
	Add(item *Item) (lastID bson.ObjectID, err error)
	Update(newItem *Item) (bool, error)
	Delete(id string) (bool, error)
}
