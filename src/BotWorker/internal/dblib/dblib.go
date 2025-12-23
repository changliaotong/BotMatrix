package dblib

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

type Row map[string]interface{}

func (r Row) GetInt(column string) (int, error) {
	v, ok := r[column]
	if !ok || v == nil {
		return 0, fmt.Errorf("列不存在: %s", column)
	}

	switch t := v.(type) {
	case int:
		return t, nil
	case int8:
		return int(t), nil
	case int16:
		return int(t), nil
	case int32:
		return int(t), nil
	case int64:
		return int(t), nil
	case uint:
		return int(t), nil
	case uint8:
		return int(t), nil
	case uint16:
		return int(t), nil
	case uint32:
		return int(t), nil
	case uint64:
		return int(t), nil
	case float32:
		return int(t), nil
	case float64:
		return int(t), nil
	case string:
		i, err := strconv.Atoi(t)
		if err != nil {
			return 0, fmt.Errorf("列 %s 转换为int失败: %w", column, err)
		}
		return i, nil
	case []byte:
		i, err := strconv.Atoi(string(t))
		if err != nil {
			return 0, fmt.Errorf("列 %s 转换为int失败: %w", column, err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("列 %s 类型不支持: %T", column, v)
	}
}

func (r Row) GetLong(column string) (int64, error) {
	v, ok := r[column]
	if !ok || v == nil {
		return 0, fmt.Errorf("列不存在: %s", column)
	}

	switch t := v.(type) {
	case int:
		return int64(t), nil
	case int8:
		return int64(t), nil
	case int16:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case int64:
		return t, nil
	case uint:
		return int64(t), nil
	case uint8:
		return int64(t), nil
	case uint16:
		return int64(t), nil
	case uint32:
		return int64(t), nil
	case uint64:
		return int64(t), nil
	case float32:
		return int64(t), nil
	case float64:
		return int64(t), nil
	case string:
		i, err := strconv.ParseInt(t, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("列 %s 转换为int64失败: %w", column, err)
		}
		return i, nil
	case []byte:
		i, err := strconv.ParseInt(string(t), 10, 64)
		if err != nil {
			return 0, fmt.Errorf("列 %s 转换为int64失败: %w", column, err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("列 %s 类型不支持: %T", column, v)
	}
}

func (r Row) GetValue(column string) (interface{}, error) {
	v, ok := r[column]
	if !ok {
		return nil, fmt.Errorf("列不存在: %s", column)
	}
	return v, nil
}

type Dialect interface {
	Placeholder(n int) string
	QuoteIdent(name string) string
}

type PostgresDialect struct{}

func (PostgresDialect) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

func (PostgresDialect) QuoteIdent(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

type MSSQLDialect struct{}

func (MSSQLDialect) Placeholder(n int) string {
	return fmt.Sprintf("@p%d", n)
}

func (MSSQLDialect) QuoteIdent(name string) string {
	if strings.HasPrefix(name, "[") && strings.HasSuffix(name, "]") {
		return name
	}
	return "[" + strings.ReplaceAll(name, "]", "]]") + "]"
}

type DB interface {
	Get(ctx context.Context, table, pkColumn string, pkValue interface{}, columns ...string) (Row, error)
	GetByKeys(ctx context.Context, table string, keys map[string]interface{}, columns ...string) (Row, error)
	Insert(ctx context.Context, table string, values map[string]interface{}) (int64, error)
	Update(ctx context.Context, table, pkColumn string, pkValue interface{}, values map[string]interface{}) (int64, error)
	UpdateByKeys(ctx context.Context, table string, keys map[string]interface{}, values map[string]interface{}) (int64, error)
	SetValue(ctx context.Context, table string, keys map[string]interface{}, column string, value interface{}) (int64, error)
	UpdatePlus(ctx context.Context, table string, keys map[string]interface{}, column string, delta interface{}) (int64, error)
	Delete(ctx context.Context, table, pkColumn string, pkValue interface{}) (int64, error)
	DeleteByKeys(ctx context.Context, table string, keys map[string]interface{}) (int64, error)
	Exec(ctx context.Context, query string, args ...interface{}) (int64, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

type Client struct {
	db      *sql.DB
	dialect Dialect
}

func NewClient(db *sql.DB, dialect Dialect) *Client {
	if dialect == nil {
		dialect = PostgresDialect{}
	}
	return &Client{
		db:      db,
		dialect: dialect,
	}
}

func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return c.db.QueryRowContext(ctx, query, args...)
}

func (c *Client) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return c.db.QueryContext(ctx, query, args...)
}

func buildGetQuery(dialect Dialect, table, pkColumn string, pkValue interface{}, columns ...string) (string, []interface{}) {
	selection := "*"
	if len(columns) > 0 {
		cols := make([]string, len(columns))
		for i, col := range columns {
			cols[i] = dialect.QuoteIdent(col)
		}
		selection = strings.Join(cols, ", ")
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = %s",
		selection,
		dialect.QuoteIdent(table),
		dialect.QuoteIdent(pkColumn),
		dialect.Placeholder(1),
	)

	return query, []interface{}{pkValue}
}

func buildInsertQuery(dialect Dialect, table string, values map[string]interface{}) (string, []interface{}, error) {
	if len(values) == 0 {
		return "", nil, fmt.Errorf("values 不能为空")
	}

	columns := make([]string, 0, len(values))
	placeholders := make([]string, 0, len(values))
	args := make([]interface{}, 0, len(values))

	index := 1
	for col, val := range values {
		columns = append(columns, dialect.QuoteIdent(col))
		placeholders = append(placeholders, dialect.Placeholder(index))
		args = append(args, val)
		index++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		dialect.QuoteIdent(table),
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	return query, args, nil
}

func buildUpdateQuery(dialect Dialect, table, pkColumn string, pkValue interface{}, values map[string]interface{}) (string, []interface{}, error) {
	if len(values) == 0 {
		return "", nil, fmt.Errorf("values 不能为空")
	}

	sets := make([]string, 0, len(values))
	args := make([]interface{}, 0, len(values)+1)

	index := 1
	for col, val := range values {
		sets = append(sets, fmt.Sprintf("%s = %s", dialect.QuoteIdent(col), dialect.Placeholder(index)))
		args = append(args, val)
		index++
	}

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = %s",
		dialect.QuoteIdent(table),
		strings.Join(sets, ", "),
		dialect.QuoteIdent(pkColumn),
		dialect.Placeholder(index),
	)

	args = append(args, pkValue)

	return query, args, nil
}

func buildDeleteQuery(dialect Dialect, table, pkColumn string) string {
	return fmt.Sprintf(
		"DELETE FROM %s WHERE %s = %s",
		dialect.QuoteIdent(table),
		dialect.QuoteIdent(pkColumn),
		dialect.Placeholder(1),
	)
}

func buildWhereByKeys(dialect Dialect, keys map[string]interface{}, startIndex int) (string, []interface{}, error) {
	if len(keys) == 0 {
		return "", nil, fmt.Errorf("keys 不能为空")
	}

	parts := make([]string, 0, len(keys))
	args := make([]interface{}, 0, len(keys))

	index := startIndex
	for col, val := range keys {
		parts = append(parts, fmt.Sprintf("%s = %s", dialect.QuoteIdent(col), dialect.Placeholder(index)))
		args = append(args, val)
		index++
	}

	return strings.Join(parts, " AND "), args, nil
}

func scanSingleRow(rows *sql.Rows) (Row, error) {
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, err
		}
		return nil, sql.ErrNoRows
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	row := Row{}
	for i, col := range columns {
		row[col] = values[i]
	}

	if rows.Next() {
		return nil, fmt.Errorf("查询返回多行")
	}

	return row, nil
}

func (c *Client) Get(ctx context.Context, table, pkColumn string, pkValue interface{}, columns ...string) (Row, error) {
	query, args := buildGetQuery(c.dialect, table, pkColumn, pkValue, columns...)
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanSingleRow(rows)
}

func (c *Client) GetByKeys(ctx context.Context, table string, keys map[string]interface{}, columns ...string) (Row, error) {
	selection := "*"
	if len(columns) > 0 {
		cols := make([]string, len(columns))
		for i, col := range columns {
			cols[i] = c.dialect.QuoteIdent(col)
		}
		selection = strings.Join(cols, ", ")
	}

	whereClause, args, err := buildWhereByKeys(c.dialect, keys, 1)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		selection,
		c.dialect.QuoteIdent(table),
		whereClause,
	)

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanSingleRow(rows)
}

func (c *Client) Insert(ctx context.Context, table string, values map[string]interface{}) (int64, error) {
	query, args, err := buildInsertQuery(c.dialect, table, values)
	if err != nil {
		return 0, err
	}
	return c.Exec(ctx, query, args...)
}

func (c *Client) Update(ctx context.Context, table, pkColumn string, pkValue interface{}, values map[string]interface{}) (int64, error) {
	query, args, err := buildUpdateQuery(c.dialect, table, pkColumn, pkValue, values)
	if err != nil {
		return 0, err
	}
	return c.Exec(ctx, query, args...)
}

func (c *Client) UpdateByKeys(ctx context.Context, table string, keys map[string]interface{}, values map[string]interface{}) (int64, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("values 不能为空")
	}
	if len(keys) == 0 {
		return 0, fmt.Errorf("keys 不能为空")
	}

	sets := make([]string, 0, len(values))
	args := make([]interface{}, 0, len(values))

	index := 1
	for col, val := range values {
		sets = append(sets, fmt.Sprintf("%s = %s", c.dialect.QuoteIdent(col), c.dialect.Placeholder(index)))
		args = append(args, val)
		index++
	}

	whereClause, whereArgs, err := buildWhereByKeys(c.dialect, keys, index)
	if err != nil {
		return 0, err
	}
	args = append(args, whereArgs...)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		c.dialect.QuoteIdent(table),
		strings.Join(sets, ", "),
		whereClause,
	)

	return c.Exec(ctx, query, args...)
}

func (c *Client) SetValue(ctx context.Context, table string, keys map[string]interface{}, column string, value interface{}) (int64, error) {
	if column == "" {
		return 0, fmt.Errorf("column 不能为空")
	}
	values := map[string]interface{}{column: value}
	return c.UpdateByKeys(ctx, table, keys, values)
}

func (c *Client) UpdatePlus(ctx context.Context, table string, keys map[string]interface{}, column string, delta interface{}) (int64, error) {
	if column == "" {
		return 0, fmt.Errorf("column 不能为空")
	}
	if len(keys) == 0 {
		return 0, fmt.Errorf("keys 不能为空")
	}

	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, delta)

	whereClause, whereArgs, err := buildWhereByKeys(c.dialect, keys, 2)
	if err != nil {
		return 0, err
	}
	args = append(args, whereArgs...)

	query := fmt.Sprintf(
		"UPDATE %s SET %s = %s + %s WHERE %s",
		c.dialect.QuoteIdent(table),
		c.dialect.QuoteIdent(column),
		c.dialect.QuoteIdent(column),
		c.dialect.Placeholder(1),
		whereClause,
	)

	return c.Exec(ctx, query, args...)
}

func (c *Client) Delete(ctx context.Context, table, pkColumn string, pkValue interface{}) (int64, error) {
	query := buildDeleteQuery(c.dialect, table, pkColumn)
	return c.Exec(ctx, query, pkValue)
}

func (c *Client) DeleteByKeys(ctx context.Context, table string, keys map[string]interface{}) (int64, error) {
	whereClause, args, err := buildWhereByKeys(c.dialect, keys, 1)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s",
		c.dialect.QuoteIdent(table),
		whereClause,
	)

	return c.Exec(ctx, query, args...)
}

type Tx struct {
	tx      *sql.Tx
	dialect Dialect
}

func (t *Tx) Exec(ctx context.Context, query string, args ...interface{}) (int64, error) {
	result, err := t.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (t *Tx) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *Tx) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *Tx) Get(ctx context.Context, table, pkColumn string, pkValue interface{}, columns ...string) (Row, error) {
	query, args := buildGetQuery(t.dialect, table, pkColumn, pkValue, columns...)
	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanSingleRow(rows)
}

func (t *Tx) GetByKeys(ctx context.Context, table string, keys map[string]interface{}, columns ...string) (Row, error) {
	selection := "*"
	if len(columns) > 0 {
		cols := make([]string, len(columns))
		for i, col := range columns {
			cols[i] = t.dialect.QuoteIdent(col)
		}
		selection = strings.Join(cols, ", ")
	}

	whereClause, args, err := buildWhereByKeys(t.dialect, keys, 1)
	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s",
		selection,
		t.dialect.QuoteIdent(table),
		whereClause,
	)

	rows, err := t.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return scanSingleRow(rows)
}

func (t *Tx) Insert(ctx context.Context, table string, values map[string]interface{}) (int64, error) {
	query, args, err := buildInsertQuery(t.dialect, table, values)
	if err != nil {
		return 0, err
	}
	return t.Exec(ctx, query, args...)
}

func (t *Tx) Update(ctx context.Context, table, pkColumn string, pkValue interface{}, values map[string]interface{}) (int64, error) {
	query, args, err := buildUpdateQuery(t.dialect, table, pkColumn, pkValue, values)
	if err != nil {
		return 0, err
	}
	return t.Exec(ctx, query, args...)
}

func (t *Tx) UpdateByKeys(ctx context.Context, table string, keys map[string]interface{}, values map[string]interface{}) (int64, error) {
	if len(values) == 0 {
		return 0, fmt.Errorf("values 不能为空")
	}
	if len(keys) == 0 {
		return 0, fmt.Errorf("keys 不能为空")
	}

	sets := make([]string, 0, len(values))
	args := make([]interface{}, 0, len(values))

	index := 1
	for col, val := range values {
		sets = append(sets, fmt.Sprintf("%s = %s", t.dialect.QuoteIdent(col), t.dialect.Placeholder(index)))
		args = append(args, val)
		index++
	}

	whereClause, whereArgs, err := buildWhereByKeys(t.dialect, keys, index)
	if err != nil {
		return 0, err
	}
	args = append(args, whereArgs...)

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s",
		t.dialect.QuoteIdent(table),
		strings.Join(sets, ", "),
		whereClause,
	)

	return t.Exec(ctx, query, args...)
}

func (t *Tx) SetValue(ctx context.Context, table string, keys map[string]interface{}, column string, value interface{}) (int64, error) {
	if column == "" {
		return 0, fmt.Errorf("column 不能为空")
	}
	values := map[string]interface{}{column: value}
	return t.UpdateByKeys(ctx, table, keys, values)
}

func (t *Tx) UpdatePlus(ctx context.Context, table string, keys map[string]interface{}, column string, delta interface{}) (int64, error) {
	if column == "" {
		return 0, fmt.Errorf("column 不能为空")
	}
	if len(keys) == 0 {
		return 0, fmt.Errorf("keys 不能为空")
	}

	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, delta)

	whereClause, whereArgs, err := buildWhereByKeys(t.dialect, keys, 2)
	if err != nil {
		return 0, err
	}
	args = append(args, whereArgs...)

	query := fmt.Sprintf(
		"UPDATE %s SET %s = %s + %s WHERE %s",
		t.dialect.QuoteIdent(table),
		t.dialect.QuoteIdent(column),
		t.dialect.QuoteIdent(column),
		t.dialect.Placeholder(1),
		whereClause,
	)

	return t.Exec(ctx, query, args...)
}

func (t *Tx) Delete(ctx context.Context, table, pkColumn string, pkValue interface{}) (int64, error) {
	query := buildDeleteQuery(t.dialect, table, pkColumn)
	return t.Exec(ctx, query, pkValue)
}

func (t *Tx) DeleteByKeys(ctx context.Context, table string, keys map[string]interface{}) (int64, error) {
	whereClause, args, err := buildWhereByKeys(t.dialect, keys, 1)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf(
		"DELETE FROM %s WHERE %s",
		t.dialect.QuoteIdent(table),
		whereClause,
	)

	return t.Exec(ctx, query, args...)
}

func (c *Client) WithTx(ctx context.Context, fn func(tx DB) error) error {
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	wrapped := &Tx{
		tx:      tx,
		dialect: c.dialect,
	}

	var fnErr error

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
		if fnErr != nil {
			_ = tx.Rollback()
		} else {
			fnErr = tx.Commit()
		}
	}()

	fnErr = fn(wrapped)

	return fnErr
}
