package clause

import "strings"

// Clause 维护当前会话的sql子句信息
type Clause struct {
	sql     map[Type]string
	sqlVars map[Type][]interface{}
}

// Type 子句类型
type Type int

const (
	INSERT Type = iota
	VALUES
	SELECT
	LIMIT
	WHERE
	ORDERBY
)

// Set 根据clause 类型和参数，设置对应的sql和sqlVars
func (c *Clause) Set(name Type, vars ...interface{}) {
	if c.sql == nil {
		c.sql = make(map[Type]string)
		c.sqlVars = make(map[Type][]interface{})
	}
	c.sql[name], c.sqlVars[name] = generators[name](vars...)
}

// Build 根据 clause子句类型，构建完整的sql和sqlVars
func (c *Clause) Build(orders ...Type) (string, []interface{}) {
	var sqls []string
	var vars []interface{}
	for _, order := range orders {
		if sql, ok := c.sql[order]; ok {
			sqls = append(sqls, sql)
			vars = append(vars, c.sqlVars[order]...)
		}
	}
	return strings.Join(sqls, " "), vars
}
