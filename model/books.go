package model

import "time"

type Book struct {
	Num              int       `json:"number"`
	Title            string    `json:"title"`
	Author           string    `json:"author"`
	Publisher        string    `json:"publisher"`
	PublicationYear  string    `json:"publicationYear"`
	ISBN             string    `json:"isbn"`
	SetISBN          string    `json:"setIsbn"`
	AdditionalCode   string    `json:"additionalCode"`
	Volume           string    `json:"volume"`
	SubjectCode      string    `json:"subjectCode"`
	BookCount        int       `json:"bookCount"`
	LoanCount        int       `json:"loanCount"`
	RegistrationDate time.Time `json:"registrationDate"`
}

type Lib struct {
	LibCode       int     `json:"libCode"`
	LibName       string  `json:"libName"`
	Address       string  `json:"address"`
	Tel           string  `json:"tel"`
	Fax           string  `json:"fax"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Homepage      string  `json:"homepage"`
	Closed        string  `json:"closed"`
	OperatingTime string  `json:"operatingTime"`
	BookCount     int     `json:"BookCount"`
}
