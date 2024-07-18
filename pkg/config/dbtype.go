//go:generate go run golang.org/x/tools/cmd/stringer -type=DBType -trimprefix DBType
package config

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type DBType int

const (
	DBTypeSQLite DBType = iota
	DBTypePostgres
	DBTypeMySQL
	DBTypeMSSQL
)

func (d *DBType) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	switch strings.ToLower(s) {
	case strings.ToLower(DBTypeSQLite.String()):
		*d = DBTypeSQLite
	case strings.ToLower(DBTypePostgres.String()):
		*d = DBTypePostgres
	case strings.ToLower(DBTypeMySQL.String()):
		*d = DBTypeMySQL
	case strings.ToLower(DBTypeMSSQL.String()):
		*d = DBTypeMSSQL
	default:
		return fmt.Errorf("invalid DBType %s", s)
	}

	return nil
}
