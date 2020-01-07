package rbac

import (
	"os"
	"path"
	"database/sql"
	// "github.com/casbin/casbin-pg-adapter"
	pgadapter "github.com/cychiuae/casbin-pg-adapter"
	"github.com/casbin/casbin/v2"

	"shortly/config"
)

func NewEnforcer(db *sql.DB, authConfig config.CasbinConfig) (*casbin.Enforcer, error) {

	a, _ := pgadapter.NewAdapter(db, "casbin")

	casbinConfig := authConfig.CasbinConfigFile
	if casbinConfig == "" {
		dir, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		casbinConfig = path.Join(dir, "app/rbac/rbac.conf")
	}

	e, err := casbin.NewEnforcer(casbinConfig, a)
	if err != nil {
		return nil, err
	}

	if err := e.LoadPolicy(); err != nil {
		return nil, err
	}

	return e, nil
}