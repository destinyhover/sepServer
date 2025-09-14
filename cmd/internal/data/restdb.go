package data

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type User struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	LastLogin int64  `json:"lastlogin"`
	Admin     int    `json:"admin"`
	Active    int    `json:"active"`
}

var dsn = os.Getenv("DATABASE_URL")

func getDbCon() *sql.DB {
	if dsn == "" {
		dsn = "postgres://admin:secret@localhost:5432/restdb?sslmode=disable" // fallback
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Println("*getDbCon():", err)
		return nil
	}
	return db
}

// FromJson decodes a serilized JSON record to User{}
func (u *User) FromJson(r io.Reader) error {
	d := json.NewDecoder(r)
	return d.Decode(u)
}

// ToJSON encodes a User to JSON
func (u *User) ToJSON(w io.Writer) error {
	e := json.NewEncoder(w)
	return e.Encode(u)
}

// FindUserID for user by id
func FindUserID(ID int64) User {
	db := getDbCon()
	if db == nil {
		log.Println("Cannot connect")
		return User{}
	}
	defer db.Close()
	rows, err := db.Query("SELECT * FROM users WHERE UserID = $1 \n", ID)
	if err != nil {
		log.Println("*FindUserID:", err)
		return User{}
	}
	var c1 int64
	var c2 string
	var c3 string
	var c4 int64
	var c5 int
	var c6 int
	u := User{}
	for rows.Next() {
		err = rows.Scan(&c1, &c2, &c3, &c4, &c5, &c6)
		if err != nil {
			log.Println("*FindUserId()", err)
			return User{}
		}
		u = User{c1, c2, c3, c4, c5, c6}
		log.Println("Found user", u)
	}
	return u
}

// FindUserName for user by username
func FindUserName(Name string) User {
	db := getDbCon()
	if db == nil {
		log.Println("Cannot connect")
		return User{}
	}
	defer db.Close()
	rows, err := db.Query("SELECT * FROM users WHERE username= $1 \n", Name)
	if err != nil {
		log.Println("*FindUserName:", err)
		return User{}
	}
	var c1 int64
	var c2 string
	var c3 string
	var c4 int64
	var c5 int
	var c6 int
	u := User{}
	for rows.Next() {
		err = rows.Scan(&c1, &c2, &c3, &c4, &c5, &c6)
		if err != nil {
			log.Println("*FindUserName()", err)
			return User{}
		}
		u = User{c1, c2, c3, c4, c5, c6}
		log.Println("Found user", u)
	}
	return u
}

// ListAllUsers() for returning all users from db with []User{}
func ListAllUsers() []User {
	db := getDbCon()
	if db == nil {
		log.Println("Cannot connect")
		return []User{}
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users\n")
	if err != nil {
		log.Println("*ListAllUsers", err)
		return []User{}
	}

	u := []User{}
	for rows.Next() {
		var us User
		err = rows.Scan(&us.ID, &us.Username, &us.Password, &us.LastLogin, &us.Admin, &us.Active)
		if err != nil {
			log.Println("*ListAllUsers()", err)
			return []User{}
		}
		u = append(u, us)
	}
	log.Println("All", u)
	return u
}

// ListAllUsers() for returning all users from db with []User{}
func ListLogged() []User {
	db := getDbCon()
	if db == nil {
		log.Println("Cannot connect")
		return []User{}
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users WHERE active = 1\n")
	if err != nil {
		log.Println("*ListLogged", err)
		return []User{}
	}

	u := []User{}
	for rows.Next() {
		var us User
		err = rows.Scan(&us.ID, &us.Username, &us.Password, &us.LastLogin, &us.Admin, &us.Active)
		if err != nil {
			log.Println("*ListLogged()", err)
			return []User{}
		}
		u = append(u, us)
	}
	log.Println("Logged", u)
	return u
}

// DeleteUser is for deleting user defined by ID
func DeleteUser(ID int64) bool {
	db := getDbCon()
	if db == nil {
		log.Println("*Cannot connect")
		return false
	}
	defer db.Close()

	t := FindUserID(ID)
	if t.ID == 0 {
		log.Println("User", ID, "doesn't exist")
		return false
	}
	stmt, err := db.Prepare("DELETE FROM users WHERE UserID = $1")
	if err != nil {
		log.Println("*db.Prepare()", err)
		return false
	}
	_, err = stmt.Exec(ID)
	if err != nil {
		log.Println("*exec", err)
		return false
	}

	return true
}

// InsertUser() is 	for adding a new user
func InsertUser(u User) bool {
	db := getDbCon()
	if db == nil {
		return false
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO users(username, password, lastlogin, admin, active) VALUES($1,$2,$3,$4,$5)")
	if err != nil {
		log.Println("*prepare", err)
		return false
	}
	_, err = stmt.Exec(u.Username, u.Password, u.LastLogin, u.Active, u.Active)
	if err != nil {
		log.Println("*exec", err)
		return false
	}
	return true
}

func IsUserAdmin(u User) bool {
	db := getDbCon()
	if db == nil {
		log.Println("Cannt con")
		return false
	}
	rows, err := db.Query("SELECT * FROM users WHERE username = $1", u.Username)
	if err != nil {
		log.Println("*Query in isuseradmin()")
		return false
	}
	var c1 int64
	var c2 string
	var c3 string
	var c4 int64
	var c5 int
	var c6 int
	t := User{}
	for rows.Next() {
		err = rows.Scan(&c1, &c2, &c3, &c4, &c5, &c6)
		if err != nil {
			log.Println("*Scan", err)
			return false
		}
		t = User{c1, c2, c3, c4, c5, c6}
	}
	if u.Username == t.Username && u.Password == t.Password && t.Admin == 1 {
		return true
	}
	return false
}

func IsUserValid(u User) bool {
	db := getDbCon()
	if db == nil {
		log.Println("Cannt con")
		return false
	}
	rows, err := db.Query("SELECT * FROM users WHERE username = $1", u.Username)
	if err != nil {
		log.Println("*Query in isuseradmin()")
		return false
	}
	var c1 int64
	var c2 string
	var c3 string
	var c4 int64
	var c5 int
	var c6 int
	t := User{}
	for rows.Next() {
		err = rows.Scan(&c1, &c2, &c3, &c4, &c5, &c6)
		if err != nil {
			log.Println("*Scan", err)
			return false
		}
		t = User{c1, c2, c3, c4, c5, c6}
	}
	if u.Username == t.Username && u.Password == t.Password {
		return true
	}
	return false
}

// UpdateUser allows you to update user name
func UpdateUser(u User) bool {
	log.Println("Updating user:", u)
	db := getDbCon()
	if db == nil {
		log.Println("Cannot con")
		return false
	}
	stmt, err := db.Prepare("UPDATE users SET username = $1, password = $2, admin =$3, active=$4 WHERE UserID = $5")
	if err != nil {
		log.Println("*Prepare upd", err)
		return false
	}
	rows, err := stmt.Exec(u.Username, u.Password, u.Admin, u.Active, u.ID)
	if err != nil {
		log.Println("*Exec upd", err)
		return false
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		log.Println("*Affected", err)
		return false
	}
	log.Println("Affected", affected)
	return true
}
