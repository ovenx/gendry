package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEq(t *testing.T) {
	var testData = []struct {
		in     map[string]any
		outCon []string
		outVal []any
	}{
		{
			map[string]any{
				"foo": "bar",
				"baz": 1,
				"qq":  "ttx",
			},
			[]string{"baz=?", "foo=?", "qq=?"},
			[]any{1, "bar", "ttx"},
		},
	}
	ass := assert.New(t)
	for _, testCase := range testData {
		cond, vals := Eq(testCase.in).Build()
		ass.Equal(len(cond), len(vals))
		ass.Equal(testCase.outCon, cond)
		ass.Equal(testCase.outVal, vals)
	}
}

func TestIn(t *testing.T) {
	var testData = []struct {
		in      map[string][]any
		outCond []string
		outVals []any
	}{
		{
			in: map[string][]any{
				"foo": {"bar", "baz"},
				"age": {5, 7, 9, 11},
			},
			outCond: []string{"age IN (?,?,?,?)", "foo IN (?,?)"},
			outVals: []any{5, 7, 9, 11, "bar", "baz"},
		},
	}
	ass := assert.New(t)
	for _, testCase := range testData {
		cond, vals := In(testCase.in).Build()
		ass.Equal(testCase.outCond, cond)
		ass.Equal(testCase.outVals, vals)
	}
}

func TestNestWhere(t *testing.T) {
	var testData = []struct {
		in      NestWhere
		outCond []string
		outVals []any
	}{
		{
			in: NestWhere([]Comparable{
				Eq(map[string]any{
					"aa": 3,
				}),
				Eq(map[string]any{
					"bb": 4,
				}),
			}),
			outCond: []string{"(aa=? AND bb=?)"},
			outVals: []any{3, 4},
		},
	}
	ass := assert.New(t)
	for _, testCase := range testData {
		cond, vals := testCase.in.Build()
		ass.Equal(testCase.outCond, cond)
		ass.Equal(testCase.outVals, vals)
	}
}

func TestResolveFields(t *testing.T) {
	ass := assert.New(t)
	m := map[string]any{
		"foo": 1,
		"bar": 2,
		"qq":  3,
		"asd": 4,
	}
	res := resolveFields(m)
	assertion := make([]string, 0)
	defaultSortAlgorithm(append(assertion, "foo", "bar", "qq", "asd"))
	for i := 0; i < len(assertion); i++ {
		ass.Equal(assertion[i], res[i])
	}
}

func TestAssembleExpression(t *testing.T) {
	var data = []struct {
		inField, inOp string
		out           string
	}{
		{"foo", "=", "foo=?"},
		{"qq", "<>", "qq<>?"},
	}
	ass := assert.New(t)
	for _, tc := range data {
		ass.Equal(tc.out, assembleExpression(tc.inField, tc.inOp))
	}
}

func TestResolveUpdate(t *testing.T) {
	var data = []struct {
		in      map[string]any
		outStr  string
		outVals []any
	}{
		{
			in: map[string]any{
				"foo": "bar",
				"bar": 1,
			},
			outStr:  "bar=?,foo=?",
			outVals: []any{1, "bar"},
		},
		{
			in: map[string]any{
				"qq":    "ttt",
				"some":  123,
				"other": 456,
			},
			outStr:  "other=?,qq=?,some=?",
			outVals: []any{456, "ttt", 123},
		},
		{
			in: map[string]any{
				"id":   1,
				"name": Raw("VALUES(name)"),
				"age":  Raw("VALUES(age)"),
			},
			outStr:  "age=VALUES(age),id=?,name=VALUES(name)",
			outVals: []any{1},
		},
		{
			in: map[string]any{
				"id":   1,
				"name": Raw("new.name"),
				"age":  Raw("new.age"),
			},
			outStr:  "age=new.age,id=?,name=new.name",
			outVals: []any{1},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		keys, vals := resolveUpdate(tc.in)
		ass.Equal(tc.outStr, keys)
		ass.Equal(tc.outVals, vals)
	}
}

func TestWhereConnector(t *testing.T) {
	var data = []struct {
		in      []Comparable
		outStr  string
		outVals []any
	}{
		{
			in: []Comparable{
				Eq(map[string]any{
					"a": "a",
					"b": "b",
				}),
				Ne(map[string]any{
					"foo": 1,
					"sex": "male",
				}),
				In(map[string][]any{
					"qq": {7, 8, 9},
				}),
			},
			outStr:  "(a=? AND b=? AND foo!=? AND sex!=? AND qq IN (?,?,?))",
			outVals: []any{"a", "b", 1, "male", 7, 8, 9},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		actualStr, actualVals := whereConnector("AND", tc.in...)
		ass.Equal(tc.outStr, actualStr)
		ass.Equal(tc.outVals, actualVals)
	}
}

func TestBuildInsert(t *testing.T) {
	var data = []struct {
		table      string
		insertType insertType
		data       []map[string]any
		outStr     string
		outVals    []any
		outErr     error
	}{
		{
			table:      "tb1",
			insertType: commonInsert,
			data: []map[string]any{
				{
					"foo": 1,
					"bar": 2,
				},
				{
					"foo": 3,
					"bar": 4,
				},
				{
					"foo": 5,
					"bar": 6,
				},
			},
			outStr:  "INSERT INTO tb1 (bar,foo) VALUES (?,?),(?,?),(?,?)",
			outVals: []any{2, 1, 4, 3, 6, 5},
			outErr:  nil,
		},
		{
			table:      "tb1",
			insertType: replaceInsert,
			data: []map[string]any{
				{
					"foo": 1,
					"bar": 2,
				},
				{
					"foo": 3,
					"bar": 4,
				},
				{
					"foo": 5,
					"bar": 6,
				},
			},
			outStr:  "REPLACE INTO tb1 (bar,foo) VALUES (?,?),(?,?),(?,?)",
			outVals: []any{2, 1, 4, 3, 6, 5},
			outErr:  nil,
		},
		{
			table:      "tb1",
			insertType: ignoreInsert,
			data: []map[string]any{
				{
					"foo": 1,
					"bar": 2,
				},
				{
					"foo": 3,
					"bar": 4,
				},
				{
					"foo": 5,
					"bar": 6,
				},
			},
			outStr:  "INSERT IGNORE INTO tb1 (bar,foo) VALUES (?,?),(?,?),(?,?)",
			outVals: []any{2, 1, 4, 3, 6, 5},
			outErr:  nil,
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		actualStr, actualVals, err := buildInsert(tc.table, tc.data, tc.insertType)
		ass.Equal(tc.outErr, err)
		ass.Equal(tc.outStr, actualStr)
		ass.Equal(tc.outVals, actualVals)
	}
}

func TestBuildInsertOnDuplicate(t *testing.T) {
	var data = []struct {
		table   string
		data    []map[string]any
		update  map[string]any
		outErr  error
		outStr  string
		outVals []any
	}{
		{
			table: "tb",
			data: []map[string]any{
				{
					"a": 1,
					"b": 2,
					"c": 3,
				},
				{
					"a": 4,
					"b": 5,
					"c": 6,
				},
			},
			update: map[string]any{
				"b": 7,
				"c": 8,
			},
			outErr:  nil,
			outStr:  "INSERT INTO tb (a,b,c) VALUES (?,?,?),(?,?,?) ON DUPLICATE KEY UPDATE b=?,c=?",
			outVals: []any{1, 2, 3, 4, 5, 6, 7, 8},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		cond, vals, err := buildInsertOnDuplicate(tc.table, tc.data, tc.update)
		ass.Equal(tc.outErr, err)
		ass.Equal(tc.outStr, cond)
		ass.Equal(tc.outVals, vals)
	}
}

func TestBuildInsertOnDuplicateMulti(t *testing.T) {
	var data = []struct {
		table   string
		data    []map[string]any
		update  []string
		outErr  error
		outStr  string
		outVals []any
	}{
		{
			table: "tb",
			data: []map[string]any{
				{
					"a": 1,
					"b": 2,
					"c": 3,
				},
				{
					"a": 4,
					"b": 5,
					"c": 6,
				},
			},
			update: []string{
				"b",
				"c",
			},
			outErr:  nil,
			outStr:  "INSERT INTO tb (a,b,c) VALUES (?,?,?),(?,?,?) ON DUPLICATE KEY UPDATE b=VALUES(b),c=VALUES(c)",
			outVals: []any{1, 2, 3, 4, 5, 6},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		cond, vals, err := buildInsertOnDuplicateMulti(tc.table, tc.data, tc.update)
		ass.Equal(tc.outErr, err)
		ass.Equal(tc.outStr, cond)
		ass.Equal(tc.outVals, vals)
	}
}

func TestBuildUpdate(t *testing.T) {
	var data = []struct {
		table      string
		conditions []Comparable
		data       map[string]any
		outErr     error
		outStr     string
		outVals    []any
	}{
		{
			table: "tb",
			conditions: []Comparable{
				Eq(map[string]any{
					"foo": "bar",
					"qq":  1,
				}),
			},
			data: map[string]any{
				"name": "deen",
				"age":  23,
			},
			outErr:  nil,
			outStr:  "UPDATE tb SET age=?,name=? WHERE (foo=? AND qq=?)",
			outVals: []any{23, "deen", "bar", 1},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		cond, vals, err := buildUpdate(tc.table, tc.data, 0, tc.conditions...)
		ass.Equal(tc.outErr, err)
		ass.Equal(tc.outStr, cond)
		ass.Equal(tc.outVals, vals)
	}
}

func TestBuildDelete(t *testing.T) {
	var data = []struct {
		table   string
		where   []Comparable
		outStr  string
		outVals []any
		outErr  error
	}{
		{
			table: "tb",
			where: []Comparable{
				Eq(map[string]any{
					"foo": 1,
					"bar": 2,
					"baz": "tt",
				}),
			},
			outStr:  "DELETE FROM tb WHERE (bar=? AND baz=? AND foo=?)",
			outVals: []any{2, "tt", 1},
			outErr:  nil,
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		actualStr, actualVals, err := buildDelete(tc.table, tc.where...)
		ass.Equal(tc.outErr, err)
		ass.Equal(tc.outStr, actualStr)
		ass.Equal(tc.outVals, actualVals)
	}
}

func TestBuildSelect(t *testing.T) {
	var data = []struct {
		table      string
		fields     []string
		conditions []Comparable
		groupBy    string
		orderBy    string
		limit      *eleLimit
		lockMode   string
		outStr     string
		outVals    []any
		outErr     error
	}{
		{
			table:  "tb",
			fields: []string{"foo", "bar"},
			conditions: []Comparable{
				Eq(map[string]any{
					"foo": 1,
					"bar": 2,
				}),
				In(map[string][]any{
					"qq": {4, 5, 6},
				}),
				OrWhere([]Comparable{
					NestWhere([]Comparable{
						Eq(map[string]any{
							"aa": 3,
						}),
						Eq(map[string]any{
							"bb": 4,
						}),
					}),
					NestWhere([]Comparable{
						Eq(map[string]any{
							"cc": 7,
						}),
						Eq(map[string]any{
							"dd": 8,
						}),
					}),
				}),
			},
			groupBy: "",
			orderBy: "foo DESC,baz ASC",
			limit: &eleLimit{
				begin: 10,
				step:  20,
			},
			lockMode: "exclusive",
			outErr:   nil,
			outStr:   "SELECT foo,bar FROM tb WHERE (bar=? AND foo=? AND qq IN (?,?,?) AND ((aa=? AND bb=?) OR (cc=? AND dd=?))) ORDER BY foo DESC,baz ASC LIMIT ?,? FOR UPDATE",
			outVals:  []any{2, 1, 4, 5, 6, 3, 4, 7, 8, 10, 20},
		},
	}
	ass := assert.New(t)
	for _, tc := range data {
		cond, vals, err := buildSelect(tc.table, tc.fields, tc.groupBy, tc.orderBy, tc.lockMode, tc.limit, tc.conditions...)
		ass.Equal(tc.outErr, err)
		ass.Equal(tc.outStr, cond)
		ass.Equal(tc.outVals, vals)
	}
}
