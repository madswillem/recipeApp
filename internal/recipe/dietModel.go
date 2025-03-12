package recipe

import "time"

type DietSchema struct {
	ID            string     `db:"id" json:"id"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	Name          string     `db:"name" json:"name"`
	Description   string     `db:"description" json:"description"`
	ExIngCategory []Category `json:"exingcategory"`
}
