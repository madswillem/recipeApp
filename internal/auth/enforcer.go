package auth

import (
	"log"

	sqlxadapter "github.com/Blank-Xu/sqlx-adapter"
	"github.com/casbin/casbin/v2"
	"github.com/jmoiron/sqlx"
)

func AddEnforcer(db *sqlx.DB) *casbin.Enforcer {
	a, err := sqlxadapter.NewAdapter(db, "casbin_rule_test")
    if err != nil {
        panic(err)
    }

    e, err := casbin.NewEnforcer("./internal/auth/model.conf", a)
    if err != nil {
        panic(err)
    }

    // Load the policy from DB.
    if err = e.LoadPolicy(); err != nil {
        log.Println("LoadPolicy failed, err: ", err)
    }

	return e
}