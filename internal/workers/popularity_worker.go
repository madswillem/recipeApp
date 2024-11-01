package workers

import (
	"sync"

	"github.com/jmoiron/sqlx"
)

type Worker struct {
	DB *sqlx.DB
}

type r_log struct {
	ID             string `db:"recipe_id"`
	Selects        int    `db:"selects"`
	Views          int    `db:"views"`
	View_change    int    `db:"view_change"`
	Selects_change int    `db:"selects_change"`
}

func (w *Worker) CreatSelectedAndViewLog(wg *sync.WaitGroup, done chan bool, err chan error) {
	println("exec")
	r := []r_log{}
	e := w.DB.Select(&r, "SELECT id as recipe_id, selects, views FROM recipes")
	if e != nil {
		print("hi")
		err <- e
		wg.Done()
		return
	}
	r_l := []r_log{}
	r_l, e = GetLastLog(w.DB)
	if e != nil {
		print("ho")
		err <- e
		wg.Done()
		return
	}

	r_n := CreateDiff(r, r_l)
	_, e = w.DB.NamedExec(`INSERT INTO recipe_selects_views_log (recipe_id, selects, views, view_change, selects_change)
		VALUES (:recipe_id, :selects, :views, :view_change, :selects_change)`, r_n)

	err <- e
	done <- true
	wg.Done()
}

func GetLastLog(db *sqlx.DB) ([]r_log, error) {
	r := []r_log{}
	err := db.Select(&r, `SELECT DISTINCT ON (recipe_id) recipe_id, selects, views, view_change, selects_change
			FROM recipe_selects_views_log
			ORDER BY recipe_id, day DESC;`)
	return r, err
}

func CreateDiff(r1, r2 []r_log) []r_log {
	var d []r_log
	for _, v1 := range r1 {
		found := false
		for _, v2 := range r2 {
			if v1.ID == v2.ID {
				d = append(d, r_log{ID: v1.ID, Selects: v1.Selects, Views: v1.Views, Selects_change: v1.Selects - v2.Selects, View_change: v1.Views - v2.Views})
				found = true
				break
			}
		}
		if !found {
			d = append(d, v1)
		}
	}

	return d
}
