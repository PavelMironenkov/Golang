package items

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
	// mgo "gopkg.in/mgo.v2"
	// BSON "gopkg.in/mgo.v2/bson"
)

var itemCollection *mongo.Collection

// go test -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html

func TestAddAndGetByID(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	sess, _ := mgo.Dial("mongodb://localhost")
	collection := sess.DB("coursera").C("items")
	repo := NewMemoryRepo(collection)
	expectedItem := Item{
		Score: 1496,
		Views: 1,
		Type:  "text",
		Title: "",
		Author: Author{Username: "",
			ID: ""},
		Category:         "",
		Text:             "ctfvygbuhnj",
		URL:              "",
		Votes:            []Vote{},
		Comments:         []Comment{},
		Created:          "",
		UpvotePercentage: 0,
	}
	mt.Run("success add and get by id", func(mt *mtest.T) {
		lastInsertID, err := repo.Add(&expectedItem)
		expectedItem.PostID = lastInsertID
		assert.NotEmpty(t, lastInsertID)
		assert.Nil(t, err)
		itemResponse, err := repo.GetByID(lastInsertID.Hex())
		assert.Equal(t, &expectedItem, itemResponse)
		assert.Nil(t, err)
	})
	mt.Run("error", func(mt *mtest.T) {
		NewMemoryRepo(mt.Coll)
		mt.Coll
		mt.AddMockResponses(bson.D{{"ok", 0}})
		
	
		repo.Add(&Item{})
	})
	// mt.AddMockResponses()
}

func TestGetByIDBadCases(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	sess, _ := mgo.Dial("mongodb://localhost")
	collection := sess.DB("coursera").C("items")
	repo := NewMemoryRepo(collection)

	mt.Run("not found", func(mt *mtest.T) {
		itemResponse, err := repo.GetByID(BSON.NewObjectId().Hex())
		assert.Error(t, err)
		assert.Empty(t, itemResponse)
	})

	mt.Run("bad id", func(mt *mtest.T) {
		itemResponse, err := repo.GetByID(BSON.NewObjectId().String())
		assert.Empty(t, itemResponse)
		assert.Error(t, err)
	})
}
