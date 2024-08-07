// Code generated by "stringer -type=DBType -trimprefix DBType"; DO NOT EDIT.

package config

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[DBTypeSQLite-0]
	_ = x[DBTypePostgres-1]
	_ = x[DBTypeMySQL-2]
	_ = x[DBTypeMSSQL-3]
}

const _DBType_name = "SQLitePostgresMySQLMSSQL"

var _DBType_index = [...]uint8{0, 6, 14, 19, 24}

func (i DBType) String() string {
	if i < 0 || i >= DBType(len(_DBType_index)-1) {
		return "DBType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _DBType_name[_DBType_index[i]:_DBType_index[i+1]]
}
