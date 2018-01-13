package dbs

import (
	"bytes"
	"strings"
	"strconv"
	"database/sql"
	"fmt"
	"errors"
)

type DeleteBuilder struct {
	prefixes     rawClauses
	options      rawClauses
	alias        []string
	tables       rawClauses
	using        string
	joins        []string
	joinsArg     []interface{}
	wheres       whereExpressions
	wheres2      Clause
	orderBys     []string
	limit        uint64
	updateLimit  bool
	offset       uint64
	updateOffset bool
	suffixes     rawClauses
}

func (this *DeleteBuilder) Prefix(sql string, args ...interface{}) *DeleteBuilder {
	this.prefixes = append(this.prefixes, SQL(sql, args...))
	return this
}

func (this *DeleteBuilder) Options(options ...string) *DeleteBuilder {
	for _, c := range options {
		this.options = append(this.options, SQL(c))
	}
	return this
}

func (this *DeleteBuilder) Alias(alias ...string) *DeleteBuilder {
	this.alias = append(this.alias, alias...)
	return this
}

func (this *DeleteBuilder) Table(table string, args ...string) *DeleteBuilder {
	var ts []string
	ts = append(ts, fmt.Sprintf("`%s`", table))
	ts = append(ts, args...)
	this.tables = append(this.tables, SQL(strings.Join(ts, " ")))
	return this
}

func (this *DeleteBuilder) USING(sql string) *DeleteBuilder {
	this.using = sql
	return this
}

//func (this *DeleteBuilder) Join(join, table string) *DeleteBuilder {
//	return this.join(join, table)
//}
//
//func (this *DeleteBuilder) RightJoin(table string) *DeleteBuilder {
//	return this.join("RIGHT JOIN", table)
//}
//
//func (this *DeleteBuilder) LeftJoin(table string) *DeleteBuilder {
//	return this.join("LEFT JOIN", table)
//}
//
//func (this *DeleteBuilder) join(join, table string) *DeleteBuilder {
//	this.joins = append(this.joins, join, fmt.Sprintf("`%s`", table))
//	return this
//}


func (this *DeleteBuilder) Join(join, table, suffix string, args ...interface{}) *DeleteBuilder {
	return this.join(join, table, suffix, args...)
}

func (this *DeleteBuilder) RightJoin(table, suffix string, args ...interface{}) *DeleteBuilder {
	return this.join("RIGHT JOIN", table, suffix, args...)
}

func (this *DeleteBuilder) LeftJoin(table, suffix string, args ...interface{}) *DeleteBuilder {
	return this.join("LEFT JOIN", table, suffix, args...)
}

func (this *DeleteBuilder) join(join, table, suffix string, args ...interface{}) *DeleteBuilder {
	this.joins = append(this.joins, join, fmt.Sprintf("`%s`", table), suffix)
	this.joinsArg = append(this.joinsArg, args...)
	return this
}

func (this *DeleteBuilder) Where2(sql string, args ...interface{}) *DeleteBuilder {
	this.wheres = append(this.wheres, WhereExpression(sql, args...))
	return this
}

func (this *DeleteBuilder) Where(c Clause) *DeleteBuilder {
	this.wheres2 = c
	return this
}

func (this *DeleteBuilder) OrderBy(sql ...string) *DeleteBuilder {
	this.orderBys = append(this.orderBys, sql...)
	return this
}

func (this *DeleteBuilder) Limit(limit uint64) *DeleteBuilder {
	this.limit = limit
	this.updateLimit = true
	return this
}

func (this *DeleteBuilder) Offset(offset uint64) *DeleteBuilder {
	this.offset = offset
	this.updateOffset = true
	return this
}

func (this *DeleteBuilder) Suffix(sql string, args ...interface{}) *DeleteBuilder {
	this.suffixes = append(this.suffixes, SQL(sql, args...))
	return this
}

func (this *DeleteBuilder) ToSQL() (sql string, args []interface{}, err error) {
	if len(this.tables) == 0 {
		return "", nil, errors.New("delete statements must specify a table")
	}

	var sqlBuffer = &bytes.Buffer{}
	if len(this.prefixes) > 0 {
		args, _ = this.prefixes.AppendToSQL(sqlBuffer, " ", args)
		sqlBuffer.WriteString(" ")
	}

	sqlBuffer.WriteString("DELETE ")

	if len(this.options) > 0 {
		args, _ = this.options.AppendToSQL(sqlBuffer, " ", args)
		sqlBuffer.WriteString(" ")
	}

	if len(this.alias) > 0 {
		sqlBuffer.WriteString(strings.Join(this.alias, ", "))
		sqlBuffer.WriteString(" ")
	}

	sqlBuffer.WriteString("FROM ")

	if len(this.tables) > 0 {
		args, _ = this.tables.AppendToSQL(sqlBuffer, ", ", args)
	}

	if len(this.using) > 0 {
		sqlBuffer.WriteString(" USING ")
		sqlBuffer.WriteString(this.using)
	}

	if len(this.joins) > 0 {
		sqlBuffer.WriteString(" ")
		sqlBuffer.WriteString(strings.Join(this.joins, " "))
		args = append(args, this.joinsArg...)
	}

	if len(this.wheres) == 0 && this.wheres2 == nil {
		return "", nil, errors.New("delete statements must have WHERE condition")
	}

	if this.wheres2 != nil {
		wSql, wArgs , _ := this.wheres2.ToSQL()
		if len(strings.TrimSpace(wSql)) > 0 {
			sqlBuffer.WriteString(" WHERE ")
			sqlBuffer.WriteString(wSql)
			args = append(args, wArgs...)
		}
	} else if len(this.wheres) > 0 {
		sqlBuffer.WriteString(" WHERE ")
		args, _ = this.wheres.appendToSQL(sqlBuffer, " ", args)
	}

	if len(this.orderBys) > 0 {
		sqlBuffer.WriteString(" ORDER BY ")
		sqlBuffer.WriteString(strings.Join(this.orderBys, ", "))
	}

	if this.updateLimit {
		sqlBuffer.WriteString(" LIMIT ")
		sqlBuffer.WriteString(strconv.FormatUint(this.limit, 10))
	}

	if this.updateOffset {
		sqlBuffer.WriteString(" OFFSET ")
		sqlBuffer.WriteString(strconv.FormatUint(this.offset, 10))
	}

	if len(this.suffixes) > 0 {
		sqlBuffer.WriteString(" ")
		args, _ = this.suffixes.AppendToSQL(sqlBuffer, " ", args)
	}

	sql = sqlBuffer.String()

	return sql, args, err
}

func (this *DeleteBuilder) Exec(s SQLExecutor) (sql.Result, error) {
	sql, args, err := this.ToSQL()
	if err != nil {
		return nil, err
	}
	return Exec(s, sql, args...)
}

func NewDeleteBuilder() *DeleteBuilder {
	return &DeleteBuilder{}
}