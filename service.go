package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-sql-driver/mysql"

	"github.com/gin-gonic/gin"
)

type User struct {
	ID       int
	FullName string
	Company  string
	Contact  string
	Comment  string
}

type UserPage struct {
	Page         int    `json:"Page"`
	PreviousPage int    `json:"PreviousPage"`
	NextPage     int    `json:"NextPage"`
	User         []User `json:"User"`
}

var db *sql.DB

func main() {
	router := gin.Default()
	cfg := mysql.Config{
		User:                 "root",
		Passwd:               "",
		AllowNativePasswords: true,
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "mytestdb",
	}
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	pingErr := db.Ping()
	if err != nil || pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected!")
	defer db.Close()

	router.GET("/users", getUsers)

	router.Run("localhost:8078")
}

func getUsersFromDB(page int) ([]User, error) {
	var users []User
	startLimit := page * 10
	endLimit := startLimit + 10
	fmt.Println(startLimit)
	fmt.Println(endLimit)
	rows, err := db.Query("SELECT * FROM users LIMIT ?,?", startLimit, endLimit)
	if err != nil {
		return nil, fmt.Errorf("there is a problem while retrieving users")
	}
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.FullName, &user.Company, &user.Contact, &user.Comment); err != nil {
			return nil, fmt.Errorf("a problem occured while parsing")
		}
		fmt.Println(user)
		users = append(users, user)
	}
	fmt.Println("Kullanicilar")
	fmt.Println(users)
	return users, nil
}

func getUsers(c *gin.Context) {
	urlValues := c.Request.URL.Query()
	fmt.Println(urlValues)
	page, _ := strconv.Atoi(urlValues.Get("Page"))
	var previousPage int
	if page < 1 {
		previousPage = 0
	} else {
		previousPage = page - 1
	}
	nextPage := page + 1
	users, err := getUsersFromDB(page)
	if err != nil {

		fmt.Println(err.Error())
	}
	fmt.Println(users)
	json := UserPage{
		User:         users,
		Page:         page,
		PreviousPage: previousPage,
		NextPage:     nextPage,
	}
	c.IndentedJSON(http.StatusOK, json)
}
