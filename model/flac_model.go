package model

type APIResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Code      int    `json:"code"`
	Result    Result `json:"result"`
	Timestamp int64  `json:"timestamp"`
}

type Result struct {
	Total int    `json:"total"`
	List  []Song `json:"list"`
}

type Song struct {
	Singers   []string `json:"singers"`
	AlbumName string   `json:"albumName"`
	PicURL    string   `json:"picUrl"`
	Name      string   `json:"name"`
	ID        string   `json:"id"`
}

type FileURLResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Code      int    `json:"code"`
	Result    string `json:"result"`
	Timestamp int64  `json:"timestamp"`
}
