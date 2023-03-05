package items

import (
	ctx "context"
	"errors"

	// mgo "gopkg.in/mgo.v2"
	// "gopkg.in/mgo.v2/bson"
	bson "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ItemMemoryRepository struct {
	items *mongo.Collection
}

func NewMemoryRepo(items *mongo.Collection) ItemMemoryRepository {
	return ItemMemoryRepository{
		items: items,
	}
}

func (repo ItemMemoryRepository) GetAll() ([]*Item, error) {
	items := []*Item{}
	// err := repo.items.Find(bson.M{}).All(&items)
	cursor, err := repo.items.Find(ctx.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	err = cursor.All(ctx.Background(), &items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (repo ItemMemoryRepository) GetByID(id string) (*Item, error) {
	if !bson.IsValidObjectID(id) {
		return nil, errors.New("bad id")
	}
	ID, _ := bson.ObjectIDFromHex(id)
	post := &Item{}
	err := repo.items.FindOne(ctx.Background(), bson.M{"_id": ID}).Decode(&post)
	if err != nil {
		return &Item{}, err
	}
	return post, nil

}

func (repo ItemMemoryRepository) Add(item *Item) (lastID bson.ObjectID, err error) {
	PostID := bson.NewObjectID()
	newItem := bson.M{
		"Score":            item.Score,
		"Views":            item.Views,
		"Type":             item.Type,
		"Title":            item.Title,
		"Author":           item.Author,
		"Category":         item.Category,
		"Text":             item.Text,
		"URL":              item.URL,
		"Votes":            item.Votes,
		"Comments":         item.Comments,
		"Created":          item.Created,
		"UpvotePercentage": item.UpvotePercentage,
		"_id":              PostID,
	}
	_, err = repo.items.InsertOne(ctx.TODO(), newItem)
	if err != nil {
		return bson.ObjectID{}, err
	}
	return PostID, nil
}

func (repo ItemMemoryRepository) Update(newItem *Item) (bool, error) {
	_, err := repo.items.UpdateOne(ctx.TODO(), bson.M{"_id": newItem.PostID}, newItem)
	if err == nil {
 		return false, err
	}
	return true, nil
}

func (repo ItemMemoryRepository) Delete(id string) (bool, error) {

	if !bson.IsValidObjectID(id) {
		return false, errors.New("bad id")
	}
	ID, _ := bson.ObjectIDFromHex(id)
	//ID := bson.ObjectIdHex(id)
	_, err := repo.items.DeleteOne(ctx.TODO(), bson.M{"_id": ID})
	if err != nil {
		return false, err
	}
	return true, nil
}
