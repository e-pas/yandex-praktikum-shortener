package model

type ShortURL struct {
	Short   string `json:"SHORT"`
	URL     string `json:"URL"`
	UserID  string `json:"USERID"`
	Deleted bool   `json:"DELETED"`
}
