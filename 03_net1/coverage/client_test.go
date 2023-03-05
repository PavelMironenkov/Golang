package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

type TestCase1 struct {
	client   SearchClient
	request  SearchRequest
	filename string
	IsError  bool
}

func ChechOut(err error) {
	if err != nil {
		panic(err)
	}
}

func ErrTestHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	res, err1 := json.Marshal(123)
	ChechOut(err1)
	_, err2 := w.Write(res)
	ChechOut(err2)
}

func TimeOutHandler(w http.ResponseWriter, r *http.Request) {
	time.Sleep(2 * time.Second)
}
func ErrUnmarshalHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	res, err1 := json.Marshal(345)
	ChechOut(err1)
	_, err2 := w.Write(res)
	ChechOut(err2)
}

func TestCheckout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	ts1 := httptest.NewServer(http.HandlerFunc(ErrTestHandler))
	ts2 := httptest.NewServer(http.HandlerFunc(TimeOutHandler))
	ts3 := httptest.NewServer(http.HandlerFunc(ErrUnmarshalHandler))
	cases := []TestCase1{
		{
			client:  SearchClient{"dewe", ts.URL},
			request: SearchRequest{-5, 1, "Name", "Name", 1},
			IsError: true,
		},
		{
			client:  SearchClient{"dewe", ts.URL},
			request: SearchRequest{26, 1, "Name", "Name", 1},
			IsError: false,
		},
		{
			client:  SearchClient{"dewe", ts.URL},
			request: SearchRequest{2, -11, "Name", "Name", 1},
			IsError: true,
		},
		{
			client:  SearchClient{"", ts.URL},
			request: SearchRequest{1, 1, "Name", "Name", 1},
			IsError: true,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{1, 1, "Be", "Id", 1},
			IsError: false,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{1, 1, "Be", "Id", -1},
			IsError: false,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{1, 1, "Be", "Id", 0},
			IsError: false,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{1, 1, "ua", "Age", 3},
			IsError: true,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{1, 1, "ua", "Age", -1},
			IsError: false,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{1, 1, "ua", "Age", 1},
			IsError: false,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{1, 1, "ua", "Age", 0},
			IsError: false,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{1, 1, "Be", "Name", 1},
			IsError: false,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{2, 1, "Name", "", -1},
			IsError: false,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{1, 1, "Name", "fd", 0},
			IsError: true,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{3, 1, "", "", -1},
			IsError: false,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{3, 1, "", "", 0},
			IsError: false,
		},
		{
			client:   SearchClient{"dewesf", ts.URL},
			request:  SearchRequest{3, 1, "Name", "", -1},
			filename: "ececded",
			IsError:  true,
		},
		{
			client:  SearchClient{"dewesf", ts3.URL},
			request: SearchRequest{3, 1, "ua", "", -1},
			IsError: true,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{1000, 0, "Be", "", 0},
			IsError: false,
		},
		{
			client:  SearchClient{"dewesf", ts1.URL},
			request: SearchRequest{1000, 0, "Be", "", 0},
			IsError: true,
		},
		{
			client:  SearchClient{"dewesf", ts2.URL},
			request: SearchRequest{1000, 0, "Be", "", 0},
			IsError: true,
		},
		{
			client:  SearchClient{"dewesf", ""},
			request: SearchRequest{1000, 0, "Be", "", 0},
			IsError: true,
		},
		{
			client:  SearchClient{"dewesf", ts.URL},
			request: SearchRequest{1000, 0, "Be", "", 1},
			filename: "dataset copy.xml",
			IsError: true,
		},
		
	}
	for caseNum, item := range cases {
		if item.filename != "" {
			Filename = item.filename
		}
		res, err := item.client.FindUsers(item.request)
		Filename = "dataset.xml"
		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		fmt.Println("[", caseNum, "] response: ", res)
	}
	ts.Close()
	ts1.Close()
	ts2.Close()
	ts3.Close()
}

type TestCase2 struct {
	client SearchClient
	IsError bool
}

func ResponseChechout(resp *http.Response) error{
	switch resp.StatusCode {
	case http.StatusBadRequest:
		return fmt.Errorf("Unknown bad request error")
	case http.StatusInternalServerError:
		return fmt.Errorf("SearchServer fatal error")
	default:
		return nil
	}
}

func ServChech(item TestCase2) error{
	searcherReq, _ := http.NewRequest("GET", item.client.URL , nil) //nolint:errcheck
	searcherReq.Header.Add("AccessToken", item.client.AccessToken)
	resp, err := client.Do(searcherReq)
	ChechOut(err)
	return ResponseChechout(resp)
}

func TestCasesErrorsParametrs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))
	cases := []TestCase2{
		{	
			client: SearchClient{"gbtr", ts.URL + "?" + "limit=a&offset=0&query=fehcbj&order_field=efncie&order_by=0"},
			IsError:    true,
		},
		{	
			client: SearchClient{"gbtr", ts.URL + "?" + "limit=1&offset=a&query=fehcbj&order_field=efncie&order_by=0"},
			IsError:    true,
		},
		{	
			client: SearchClient{"gbtr", ts.URL + "?" + "limit=1&offset=0&query=fehcbj&order_field=efncie&order_by=a"},
			IsError:    true,
		},
	}
	for caseNum, item := range cases {
		err := ServChech(item)
		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if err != nil && item.IsError{
			fmt.Printf("[%d] expected error, got error: %s\n", caseNum, err)
		}
	}
}

type TestCase3 struct{
	nameFunc string
	IsError bool
}

func ErrorUnavailableOnServerHandler(w http.ResponseWriter, r *http.Request){
	name := r.FormValue("name")
	caseNum := r.FormValue("caseNum")
	switch name{
	case "marshal":
		value := make(chan int)
		Marshal(w, value)
	case "decoder":
		file := OpenFile("client.go")
		decoder := xml.NewDecoder(file)
		DecodeElement(w, decoder, nil)
	case "write":
		res := "ghfshjgkfndfhbjhdbvfпааврориорв kbdjfhbs"
		if !Write(w, []byte(res), 2) {
			fmt.Printf("[%s] expected error, got error: %s\n", caseNum, "SearchServer fatal error" )
		}
	}
}

func TestErrorUnavailableOnServer(t *testing.T){
	ts := httptest.NewServer(http.HandlerFunc(ErrorUnavailableOnServerHandler))
	cases := []TestCase3{
		{
			nameFunc:	"decoder",
			IsError: true,
		},
		{
			nameFunc:	"marshal",
			IsError: true,
		},
		{
			nameFunc:	"write",
			IsError: false,
		},
	}

	for caseNum, item := range cases {
		searcherReq, _ := http.NewRequest("GET", ts.URL + "?name=" + item.nameFunc + "&caseNum=" + strconv.Itoa(caseNum), nil) //nolint:errcheck
		resp, err := client.Do(searcherReq)
		ChechOut(err)
		err = ResponseChechout(resp)
		if err != nil && !item.IsError {
			t.Errorf("[%d] unexpected error: %#v", caseNum, err)
		}
		if err == nil && item.IsError {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if err != nil && item.IsError{
			fmt.Printf("[%d] expected error, got error: %s\n", caseNum, err)
		}
	}
}
