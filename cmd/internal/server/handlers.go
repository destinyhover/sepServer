package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	_ "restapi/cmd/internal/data"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type errorResp struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResp{Error: msg})
	log.Printf("ERROR %d: %s", status, msg)
}

// SliceToJSON encodes slice with JSON records
func SliceToJSON(slice interface{}, w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(slice)
}

type notAllowedHandler struct{}

func (n notAllowedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	MethodNotAllowedHandler(w, r)
}

// MethodNotAllowedHandler is executed when the HTTP method is incorrect
func MethodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("405:", r.Method, r.URL.Path)
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

// MethodNotAllowedHandler is execute when method is incorrect
func DefaultHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("404:", r.Method, r.URL.Path)
	writeError(w, http.StatusNotFound, fmt.Sprintf("path %s is not supported", r.URL.Path))
}

// TimeHandler is for handling /time
func TimeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving TimeHandler", r.URL.Path, r.Host)
	now := time.Now().Format(time.RFC1123)
	writeJSON(w, http.StatusOK, map[string]string{"time": now})
}

// AddHandler is for adding a new user
// Accept 2 users, [0] is admin, [2] is to add
func AddHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("AddHandler serving", r.URL.Path, r.Host)

	d, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cannot read body")
		return
	}
	if len(d) == 0 {
		writeError(w, http.StatusBadRequest, "empty body")
		return
	}

	var users []User
	if err := json.Unmarshal(d, &users); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: expected array [admin, user]")
		return
	}
	if len(users) < 2 {
		writeError(w, http.StatusBadRequest, "expected two users: admin and user-to-add")
		return
	}

	if !IsUserAdmin(users[0]) {
		writeError(w, http.StatusForbidden, "admin required")
		return
	}

	if ok := InsertUser(users[1]); !ok {
		writeError(w, http.StatusBadRequest, "failed to insert user")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving DeleteHandler", r.URL.Path, r.Host)

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	cId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var admin User
	if err := admin.FromJson(r.Body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !IsUserAdmin(admin) {
		writeError(w, http.StatusForbidden, "admin required")
		return
	}

	t := FindUserID(cId)
	if t.ID == 0 {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	if !DeleteUser(cId) {
		writeError(w, http.StatusBadRequest, "failed to delete user")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// GetAllHandler is for getting all data users from db
func GetAllHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving GetAllHandler", r.URL.Path, r.Host)

	d, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cannot read body")
		return
	}
	if len(d) == 0 {
		writeError(w, http.StatusBadRequest, "empty body")
		return
	}

	var admin User
	if err := json.Unmarshal(d, &admin); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !IsUserAdmin(admin) {
		writeError(w, http.StatusForbidden, "admin required")
		return
	}

	list := ListAllUsers()
	writeJSON(w, http.StatusOK, list)
}

// GetLoggedHandler is for getting all data logged users from db
func GetLoggedHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving GetLoggedHandler", r.URL.Path, r.Host)

	d, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "cannot read body")
		return
	}
	if len(d) == 0 {
		writeError(w, http.StatusBadRequest, "empty body")
		return
	}

	var admin User
	if err := json.Unmarshal(d, &admin); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !IsUserAdmin(admin) {
		writeError(w, http.StatusForbidden, "admin required")
		return
	}

	list := ListLogged()
	writeJSON(w, http.StatusOK, list)
}

func GetIdHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving GetIdHandler", r.URL.Path, r.Host)

	username, ok := mux.Vars(r)["username"]
	if !ok || username == "" {
		writeError(w, http.StatusBadRequest, "missing username")
		return
	}

	var admin User
	if err := admin.FromJson(r.Body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !IsUserAdmin(admin) {
		writeError(w, http.StatusForbidden, "admin required")
		return
	}

	t := FindUserName(username)
	if t.Username == "" {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// GetUserDataHandler + GET returns full record of a user
func GetUserDataHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving GetUserDataHandler", r.URL.Path, r.Host)

	idStr, ok := mux.Vars(r)["id"]
	if !ok {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	cId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var admin User
	if err := admin.FromJson(r.Body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !IsUserAdmin(admin) {
		writeError(w, http.StatusForbidden, "admin required")
		return
	}

	t := FindUserID(cId)
	if t.ID == 0 {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

// UpdateHandler is for getting upd data of existig user + PUT
func UpdateHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving UpdateHandler", r.URL.Path, r.Host)

	d, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		writeError(w, http.StatusBadRequest, "cannot read body")
		return
	}
	if len(d) == 0 {
		log.Println(err)
		writeError(w, http.StatusBadRequest, "empty body")
		return
	}

	var in []User
	if err := json.Unmarshal(d, &in); err != nil {
		log.Println(err)
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}

	// В текущей схеме in содержит и админа, и обновляемого пользователя — это неочевидно.
	// Если у тебя один объект, нужно отдельно передавать данные админа.
	// Ниже — как у тебя было: проверка admin по тем же полям (НЕ рекомендуем так делать).
	if !IsUserAdmin(in[0]) {
		writeError(w, http.StatusForbidden, "admin required")
		return
	}

	out := FindUserID(in[1].ID)
	if out.ID == 0 {
		log.Println(err)
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	out.Username = in[1].Username
	out.Password = in[1].Password
	out.Admin = in[1].Admin

	if !UpdateUser(out) {
		writeError(w, http.StatusBadRequest, "failed to update")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// LoginHandler  is for upd LastLogin and change Active
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving LoginHandler", r.URL.Path, r.Host)

	var in User
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !IsUserValid(in) {
		writeError(w, http.StatusBadRequest, "invalid credentials")
		return
	}

	u := FindUserName(in.Username)
	if u.ID == 0 {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	u.Active = 1
	u.LastLogin = time.Now().Unix()

	if !UpdateUser(u) {
		writeError(w, http.StatusBadRequest, "failed to update")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged-in"})
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Serving LogoutHandler", r.URL.Path, r.Host)

	var in User
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if !IsUserValid(in) {
		writeError(w, http.StatusBadRequest, "invalid credentials")
		return
	}

	u := FindUserName(in.Username)
	if u.ID == 0 {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	u.Active = 0

	if !UpdateUser(u) {
		writeError(w, http.StatusBadRequest, "failed to update")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged-out"})
}
