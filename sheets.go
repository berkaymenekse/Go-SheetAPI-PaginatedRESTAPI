package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/go-sql-driver/mysql"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type User struct {
	FullName string
	Company  string
	Contact  string
	Comment  string
}

var users []User

func getClient(config *oauth2.Config) *http.Client {
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)
	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func main() {

	cfg := mysql.Config{
		User:                 "{USERNAME}",
		Passwd:               "{PASSWORD}",
		AllowNativePasswords: true,
		Net:                  "tcp",
		Addr:                 "{ADDRESS}",
		DBName:               "{DBNAME}",
	}
	var err error
	db, err := sql.Open("mysql", cfg.FormatDSN())
	pingErr := db.Ping()
	if err != nil || pingErr != nil {
		log.Fatal(pingErr)
	}

	fmt.Println("Connected!")

	defer db.Close()

	tryRowSprintf := func(format string, row []interface{}, index int, fallback string) string {
		if index <= 0 || len(row) <= index {
			return fallback
		}
		return fmt.Sprintf(format, row[index])
	}

	countOfFilledCells := 1
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}
	spreadsheetId := "1loR_i7d4NvUYvRv9SwCL-FsmV9c2OgefGGJamejCcUs"
	readRange := "A2:E"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	if len(resp.Values) == 0 {
		fmt.Println("No data found.")
	} else {
		for _, row := range resp.Values {
			//fmt.Printf("The length of row %d, the cap of row %d", len(row), cap(row))

			user := User{}
			if len(row) > 2 {
				user.FullName = tryRowSprintf("%s", row, 1, "")
				user.Company = tryRowSprintf("%s", row, 2, "")
				user.Contact = tryRowSprintf("%s", row, 3, "")
				user.Comment = tryRowSprintf("%s", row, 4, "")
				//fmt.Println(user)
				users = append(users, user)
				countOfFilledCells += 1
				/*
					stmt, e := db.Prepare("INSERT INTO users(FullName,Company,Contact,Comment) values (?,?,?,?)")
					if e != nil {
						fmt.Printf("Unable to prepeare query: %v", err)
					}
					res, e := stmt.Exec(user.FullName, user.Company, user.Contact, user.Comment)
					if e != nil {
						log.Printf("Unable to execute query: %v", err)
					}
					fmt.Printf("res: %v\n", res)
				*/
			}
		}
		fmt.Printf("Count of all users is %d", countOfFilledCells)
	}
	fmt.Println(users)

}
