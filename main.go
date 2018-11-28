package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"time"
)

const CLIENT_ID string = "###CLIENT ID###"
const CLIENT_SECRET string = "###CLIENT SECRET###"
const LOG_PATH string = "tracksaver.log"

type authForm struct {
	ClientID string
	Callback string
	Scope    string
	State    string
}

var logger *log.Logger

var tmpl *template.Template

var AccessToken string
var TokenType string
var RefreshToken string
var ExpirationTime int64

var Client *http.Client

func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	data := authForm{CLIENT_ID,
		"http://localhost:3001/callback",
		"user-library-modify",
		randSeq(16)}
	tmpl.ExecuteTemplate(w, "index.html", data)
}

func callback(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if err := p.ByName("error"); err != "" {
		fmt.Fprintln(w, err)
		return
	}

	code := r.FormValue("code")
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", "http://localhost:3001/callback")
	data.Set("client_id", CLIENT_ID)
	data.Set("client_secret", CLIENT_SECRET)
	resp, err := http.PostForm("https://accounts.spotify.com/api/token", data)

	if err != nil {
		fmt.Fprintln(w, err)
		return
	}

	if resp.StatusCode != 200 {
		fmt.Fprintln(w, resp.Status)
		io.Copy(w, resp.Body)
		resp.Body.Close()
		return
	}

	dec := json.NewDecoder(resp.Body)
	var auth map[string]interface{}
	if err := dec.Decode(&auth); err != io.EOF && err != nil {
		fmt.Fprintln(w, err)
	}
	resp.Body.Close()
	fmt.Fprintln(w, auth)
	var ok1, ok2, ok3, ok4 bool
	var expiration float64
	AccessToken, ok1 = auth["access_token"].(string)
	TokenType, ok2 = auth["token_type"].(string)
	RefreshToken, ok3 = auth["refresh_token"].(string)
	expiration, ok4 = auth["expires_in"].(float64)
	if !(ok1 && ok2 && ok3 && ok4) {
		fmt.Fprintln(w, "Failed to retrieve tokens")
		fmt.Fprintf(w, "access:%v type:%v refresh:%v expiration:%v\n", ok1, ok2, ok3, ok4)
		return
	}
	ExpirationTime = time.Now().Add(time.Duration(expiration)*time.Second).Unix() - 30
}

func Refresh() {
	logger.Println("Refreshing token")
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", RefreshToken)
	data.Set("client_id", CLIENT_ID)
	data.Set("client_secret", CLIENT_SECRET)
	resp, err := http.PostForm("https://accounts.spotify.com/api/token", data)

	if err != nil {
		logger.Println(err)
		return
	}

	if resp.StatusCode != 200 {
		logger.Println(resp.Status)
		rstr, _ := ioutil.ReadAll(resp.Body)
		logger.Println(string(rstr))
		resp.Body.Close()
		return
	}

	dec := json.NewDecoder(resp.Body)
	var auth map[string]interface{}
	if err := dec.Decode(&auth); err != io.EOF && err != nil {
		fmt.Println(err)
		return
	}
	resp.Body.Close()
	var ok1, ok2, ok3 bool
	var expiration float64
	AccessToken, ok1 = auth["access_token"].(string)
	TokenType, ok2 = auth["token_type"].(string)
	expiration, ok3 = auth["expires_in"].(float64)
	if !(ok1 && ok2 && ok3) {
		return
	}
	ExpirationTime = time.Now().Add(time.Duration(expiration)*time.Second).Unix() - 30
}

func addSong(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	if ExpirationTime < time.Now().Unix() {
		Refresh()
	}
	data := make(map[string][]string)
	data["ids"] = []string{r.FormValue("id")}
	dataBytes, _ := json.Marshal(data)
	req, _ := http.NewRequest("PUT", "https://api.spotify.com/v1/me/tracks", bytes.NewReader(dataBytes))
	fmt.Fprintln(w, string(dataBytes))
	req.Header.Add("Authorization", TokenType+" "+AccessToken)
	resp, _ := Client.Do(req)
	if resp.StatusCode != 200 {
		rstr, _ := ioutil.ReadAll(resp.Body)
		logger.Println(string(rstr))
	} else {
		io.Copy(w, resp.Body)
	}
	resp.Body.Close()
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		c, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		b[i] = letters[c.Int64()]
	}
	return string(b)
}

func main() {

	logf, _ := os.OpenFile(LOG_PATH, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)

	prefix := fmt.Sprintf("tracksaver [%d]: ", os.Getpid())
	logger = log.New(logf, prefix, log.LstdFlags)

	Client = &http.Client{}

	router := httprouter.New()

	tmpl = template.Must(template.ParseFiles("index.html"))

	router.GET("/", index)
	router.GET("/callback", callback)
	router.POST("/addSong", addSong)

	logger.Fatal(http.ListenAndServe("127.0.0.1:3001", router))
}
