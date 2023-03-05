package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Row struct {
	ID         int    `xml:"id"`
	FirstName  string `xml:"first_name"`
	SecondName string `xml:"last_name"`
	Age        int    `xml:"age"`
	About      string `xml:"about"`
	Gender     string `xml:"gender"`
}

func OpenFile(str string) *os.File {
	file, err := os.Open(str)
	if err != nil {
		return nil
	}
	return file
}

func Marshal(w http.ResponseWriter, v interface{}) []byte {
	res, err := json.Marshal(v)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	return res
}

func Write(w http.ResponseWriter, res []byte, contentLength int) bool {
	w.Header().Add("Content-Length", strconv.Itoa(contentLength))
	_, err := w.Write(res)
	if err != nil {
		return false
	}
	w.Header().Del("Content-Length")
	return true
}

func DecodeElement(w http.ResponseWriter, decoder *xml.Decoder, tp *xml.StartElement) Row {
	var b Row
	err := decoder.DecodeElement(&b, tp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return Row{}
	}
	return b
}

func Sorting(w http.ResponseWriter, orderField string, orderBy int, resUsers []User) bool {
	switch {
	case orderField == "Name" || orderField == "":
		sort.Slice(resUsers, func(i, j int) bool {
			if orderBy == OrderByAsc {
				return resUsers[i].Name < resUsers[j].Name
			}
			if orderBy == OrderByDesc {
				return resUsers[i].Name > resUsers[j].Name
			}
			return false
		})
		w.WriteHeader(http.StatusOK)
		return true
	case orderField == "Age":
		sort.Slice(resUsers, func(i, j int) bool {
			if orderBy == OrderByAsc {
				return resUsers[i].Age < resUsers[j].Age
			}
			if orderBy == OrderByDesc {
				return resUsers[i].Age > resUsers[j].Age
			}
			return false
		})
		w.WriteHeader(http.StatusOK)
		return true
	case orderField == "Id":
		sort.Slice(resUsers, func(i, j int) bool {
			if orderBy == OrderByAsc {
				return resUsers[i].ID < resUsers[j].ID
			}
			if orderBy == OrderByDesc {
				return resUsers[i].ID > resUsers[j].ID
			}
			return false
		})
		w.WriteHeader(http.StatusOK)
		return true
	default:
		w.WriteHeader(http.StatusBadRequest)
		res := Marshal(w, SearchErrorResponse{`OrderField invalid`})
		Write(w, res, len(res))
		return false
	}
}

func CorrectParam(w http.ResponseWriter, err error) bool {
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return false
	}
	return true
}

func ParsingXML(w http.ResponseWriter) []User {
	file := OpenFile(Filename)
	if file == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil
	}
	defer file.Close()
	decoder := xml.NewDecoder(file)
	users := make([]User, 0)
	for {
		tok, err := decoder.Token()
		if tok == nil && err == io.EOF { // eof
			break
		}
		switch tp := tok.(type) {
		case xml.StartElement:
			if tp.Name.Local == "row" {
				// Декодирование элемента в структуру
				var u User
				b := DecodeElement(w, decoder, &tp)
				u.About = b.About
				u.Age = b.Age
				u.Gender = b.Gender
				u.ID = b.ID
				u.Name = b.FirstName + " " + b.SecondName
				users = append(users, u)
			}
		default:
		}
	}
	return users
}

var Filename = "dataset.xml"

func SearchServer(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Content-Length", strconv.Itoa(1))
	users := ParsingXML(w)
	if users == nil {
		return
	}
	resUsers := make([]User, 0)
	accessToken := r.Header.Get("AccessToken")
	if accessToken != "" {
		limit, err := strconv.Atoi(r.FormValue("limit"))
		if !CorrectParam(w, err) {
			return
		}
		offset, err := strconv.Atoi(r.FormValue("offset"))
		if !CorrectParam(w, err) {
			return
		}
		query := r.FormValue("query")
		orderField := r.FormValue("order_field")
		orderBy, err := strconv.Atoi(r.FormValue("order_by"))
		if !CorrectParam(w, err) {
			return
		}
		if orderBy != OrderByAsc && orderBy != OrderByDesc && orderBy != OrderByAsIs {
			w.WriteHeader(http.StatusBadRequest)
			res := Marshal(w, SearchErrorResponse{`OrderBy invalid`})
			Write(w, res, len(res))
			return
		}
		if query == "" {
			resUsers = users
		} else {
			for _, curUser := range users {
				if strings.Contains(curUser.Name, query) || strings.Contains(curUser.About, query) {
					resUsers = append(resUsers, curUser)
				}
			}
		}
		if Sorting(w, orderField, orderBy, resUsers) { // sort
			endOfSlice := offset + limit
			if offset >= len(resUsers) {
				resUsers = resUsers[:0]
				result := Marshal(w, resUsers)
				Write(w, result, len(result))
				return
			} else if endOfSlice > len(resUsers) {
				endOfSlice = len(resUsers)
			}
			resUsers = resUsers[offset:endOfSlice]
			result := Marshal(w, resUsers)
			Write(w, result, len(result))
		}
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Println(w.Header())
	}
}
