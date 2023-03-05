package handlers

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"hw5/pkg/items"
	"hw5/pkg/session"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type ItemsHandler struct {
	ItemsRepo items.ItemsRepo
	Logger    *zap.SugaredLogger
}

func MarshallandWrite(w http.ResponseWriter, data interface{}) {
	resp, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Marshalling err", http.StatusBadRequest)
		return
	}
	_, err = w.Write(resp)
	if err != nil {
		http.Error(w, "Writing response err", http.StatusInternalServerError)
	}
}

func (h *ItemsHandler) List(w http.ResponseWriter, r *http.Request) {
	elems, err := h.ItemsRepo.GetAll()
	if err != nil {
		http.Error(w, `List error: DB err - GetAll`, http.StatusInternalServerError)
		return
	}
	sort.Slice(elems, func(i, j int) bool {
		return elems[i].Score > elems[j].Score
	})
	MarshallandWrite(w, elems)
}

func (h *ItemsHandler) Add(w http.ResponseWriter, r *http.Request) {
	item := new(items.Item)
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `Bad request`, http.StatusBadRequest)
	}
	defer r.Body.Close()
	err = json.Unmarshal(bytes, item)
	if err != nil || (item.URL != "" && item.Text != "") {
		http.Error(w, `Bad form`, http.StatusBadRequest)
		return
	}

	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		http.Error(w, "Session err", http.StatusBadRequest)
		return
	}
	item.Score = 1
	item.Author.ID = strconv.Itoa(int(sess.UserID))
	item.Author.Username = sess.Login
	item.Comments = make([]items.Comment, 0)
	item.Votes = append(item.Votes, items.Vote{UserID: item.Author.ID, Vote: 1})
	item.UpvotePercentage = 100
	item.Created = time.Now().Format("2006-01-02T15:04:05.000")

	lastID, err := h.ItemsRepo.Add(item)
	if err != nil {
		http.Error(w, `Add error: DB err - Add`, http.StatusInternalServerError)
		return
	}
	MarshallandWrite(w, item)
	h.Logger.Infof("Add new post, LastInsertPostId: %v", lastID)
}

func (h *ItemsHandler) ListPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	item, err := h.ItemsRepo.GetByID(vars["POST_ID"])
	if err != nil {
		http.Error(w, `ListPost error: DB err - GetByID`, http.StatusInternalServerError)
		return
	}
	item.Views++
	flag, err := h.ItemsRepo.Update(&item)
	if err != nil && !flag {
		http.Error(w, `ListPost error: DB err - Update`, http.StatusInternalServerError)
	}
	MarshallandWrite(w, item)
	h.Logger.Infof("View post with ID: %v", item.PostID)
}

func (h *ItemsHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		http.Error(w, "You aren't authorize", http.StatusBadRequest)
		return
	}

	item, err := h.ItemsRepo.GetByID(postID)
	if err != nil {
		http.Error(w, `AddComment error: DB err - GetByID`, http.StatusInternalServerError)
		return
	}

	randID := make([]byte, 16)
	_, err = rand.Read(randID)
	if err != nil {
		http.Error(w, `Error generation ID for comment`, http.StatusInternalServerError)
	}
	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `Bad request`, http.StatusBadRequest)
	}
	defer r.Body.Close()
	comments := make(map[string]string)

	err = json.Unmarshal(bytes, &comments)
	if err != nil {
		http.Error(w, `Bad form`, http.StatusBadRequest)
		return
	}

	comm := items.Comment{
		Created:   time.Now().Format("2006-01-02T15:04:05.000"),
		Comment:   comments["comment"],
		CommentID: fmt.Sprintf("%x", randID),
		Author:    items.Author{ID: strconv.Itoa(int(sess.UserID)), Username: sess.Login},
	}
	item.Comments = append(item.Comments, comm)
	flag, err := h.ItemsRepo.Update(&item)
	if err != nil && !flag {
		http.Error(w, `AddComment error: DB err - Update`, http.StatusInternalServerError)
	}
	MarshallandWrite(w, item)
	h.Logger.Infof("Insert new comment with ID: %v, at post with ID: %v", comm.CommentID, item.PostID)
}

func (h *ItemsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]
	commID := vars["COMMENT_ID"]
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		http.Error(w, "You aren't authorize", http.StatusBadRequest)
		return
	}
	item, err := h.ItemsRepo.GetByID(postID)
	if err != nil {
		http.Error(w, `DeleteComment error: DB err - GetByID`, http.StatusInternalServerError)
		return
	}
	for i, com := range item.Comments {
		if com.CommentID == commID && com.Author.Username == sess.Login && com.Author.ID == strconv.Itoa(int(sess.UserID)) {
			item.Comments[i] = item.Comments[len(item.Comments)-1]
			// item.Comments[len(item.Comments)-1] = nil // or the zero value of T
			item.Comments = item.Comments[:len(item.Comments)-1]
			h.Logger.Infof("Delete comment with ID: %v, at post with ID: %v", com.CommentID, item.PostID)
		}
	}
	flag, err := h.ItemsRepo.Update(&item)
	if err != nil && !flag {
		http.Error(w, `DeleteComment error: DB err - Update`, http.StatusInternalServerError)
	}
	MarshallandWrite(w, item)
}

func (h *ItemsHandler) Vote(w http.ResponseWriter, r *http.Request) {
	_, strVote, _ := strings.Cut(r.URL.Path, "/api/post/")
	_, strVote, _ = strings.Cut(strVote, "/")
	vars := mux.Vars(r)
	item, err := h.ItemsRepo.GetByID(vars["POST_ID"])
	if err != nil {
		http.Error(w, `Vote error: DB err - GetByID`, http.StatusInternalServerError)
		return
	}
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		http.Error(w, "You aren't authorize", http.StatusBadRequest)
		return
	}

	userIDStr := strconv.Itoa(int(sess.UserID))
	var flag bool
	var upvote, sumVotes int
	var vote int8
	var score int64
	if strVote == "upvote" {
		vote = 1
	} else { // downvote
		vote = -1
	}
	for i, v := range item.Votes {
		if v.UserID == userIDStr {
			item.Votes[i].Vote = vote
			flag = true
		}
		if item.Votes[i].Vote == 1 {
			upvote++
		}
		score += int64(item.Votes[i].Vote)
		sumVotes++
	}
	if !flag {
		item.Votes = append(item.Votes, items.Vote{UserID: userIDStr, Vote: vote})
		score += int64(vote)
		if vote == 1 {
			upvote++
		}
		sumVotes++
	}
	item.Score = score
	if sumVotes != 0 {
		item.UpvotePercentage = uint8(100 * upvote / sumVotes)
	}else{
		item.UpvotePercentage = 0
	}
	flag, err = h.ItemsRepo.Update(&item)
	if err != nil && !flag {
		http.Error(w, `Vote error: DB err - Update`, http.StatusInternalServerError)
	}
	MarshallandWrite(w, item)
	h.Logger.Infof("Add new reaction: %v at post with ID: %v for user with ID: %v", strVote, item.PostID, sess.UserID)
}

func (h *ItemsHandler) Unvote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	item, err := h.ItemsRepo.GetByID(vars["POST_ID"])
	if err != nil {
		http.Error(w, `Vote error: DB err - GetByID`, http.StatusInternalServerError)
		return
	}
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		http.Error(w, "You aren't authorize", http.StatusBadRequest)
		return
	}

	userIDStr := strconv.Itoa(int(sess.UserID))
	var flag bool
	var upvote, sumVotes int
	var score int64
	for i, v := range item.Votes {
		if v.UserID == userIDStr {
			item.Votes[i] = item.Votes[len(item.Votes)-1]
			item.Votes = item.Votes[:len(item.Votes)-1]
			score -= int64(v.Vote)
			if v.Vote == 1{
				upvote--
			}
			sumVotes--
		}
		if v.Vote == 1 {
			upvote++
		}
		score += int64(v.Vote)
		sumVotes++
	}
	item.Score = score
	if sumVotes != 0 {
		item.UpvotePercentage = uint8(100 * upvote / sumVotes)
	}else{
		item.UpvotePercentage = 0
	}
	flag, err = h.ItemsRepo.Update(&item)
	if err != nil && !flag {
		http.Error(w, `Unvote error: DB err - Update`, http.StatusInternalServerError)
	}
	MarshallandWrite(w, item)
	h.Logger.Infof("Delete reaction user with ID: %v at post with ID: %v", sess.UserID, item.PostID)
}

func (h *ItemsHandler) GetAllAtTheCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["CATEGORY_NAME"]
	elems, err := h.ItemsRepo.GetAll()
	if err != nil {
		http.Error(w, `GetAllAtTheCategory: DB err - GetAll`, http.StatusInternalServerError)
		return
	}
	needElems := make([]items.Item, 0)
	for _, v := range elems {
		if v.Category == category {
			needElems = append(needElems, v)
		}
	}
	sort.Slice(needElems, func(i, j int) bool {
		return needElems[i].Score > needElems[j].Score
	})
	MarshallandWrite(w, needElems)
	h.Logger.Infof("Viewed all posts at category: %v", category)
}

func (h *ItemsHandler) GetAllAtUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userLogin := vars["USER_LOGIN"]
	elems, err := h.ItemsRepo.GetAll()
	if err != nil {
		http.Error(w, `GetAllAtUser error: DB err - GetAll`, http.StatusInternalServerError)
		return
	}
	needElems := make([]items.Item, 0)
	for _, v := range elems {
		if v.Author.Username == userLogin {
			needElems = append(needElems, v)
		}
	}
	sort.Slice(needElems, func(i, j int) bool {
		return needElems[i].Score > needElems[j].Score
	})
	MarshallandWrite(w, needElems)
	h.Logger.Infof("Viewed all user's posts with Login: %v", userLogin)
}

func (h *ItemsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]
	item, err := h.ItemsRepo.GetByID(postID)
	if err != nil {
		http.Error(w, `Delete error: DB err - GetByID`, http.StatusInternalServerError)
		return
	}
	sess, err := session.SessionFromContext(r.Context())
	if err != nil {
		http.Error(w, "Session err", http.StatusBadRequest)
		return
	}
	if strconv.Itoa(int(sess.UserID)) == item.Author.ID {
		ok, err := h.ItemsRepo.Delete(postID)
		if err != nil {
			http.Error(w, `Delete error: DB err - Delete`, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-type", "application/json")
		respJSON, err := json.Marshal(map[string]bool{
			"success": ok,
		})
		if err != nil {
			http.Error(w, "Marshalling err", http.StatusBadRequest)
			return
		}
		_, err = w.Write(respJSON)
		if err != nil {
			http.Error(w, "Writing response err", http.StatusInternalServerError)
		}
		h.Logger.Infof("Delete post with ID: %v for his creator-user with ID: %v", postID, sess.UserID)
	} else {
		http.Error(w, "The post was not deleted by its creator", http.StatusBadRequest)
	}
}
