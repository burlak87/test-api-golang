package domain

type Diploma struct {
	ID          int64    `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Student     *Student `json:"student"`
}