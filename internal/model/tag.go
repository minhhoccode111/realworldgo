package model

type Tag struct {
	Id       string `json:"id"`
	Name     string `json:"name"`
	Articles uint   `json:"articles"`
}
