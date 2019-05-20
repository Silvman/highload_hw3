package models

type Score struct {
	Score int `db:"score" json:"score"`
}

type Upgrade struct {
	Id          int    `db:"uid" json:"id"`
	Name        string `db:"name" json:"name"`
	Cost        int    `db:"cost" json:"cost"`
	Type        int    `db:"type" json:"type"`
	Modificator int    `db:"modificator" json:"modificator"`
	Image       string `db:"image" json:"image"`
	Duration    int    `db:"time" json:"time"`
}
