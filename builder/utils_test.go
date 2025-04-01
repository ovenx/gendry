package builder

import (
	"context"
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestResultResolver(t *testing.T) {
	var testData = []struct {
		origin   any
		intout   int64
		floatout float64
	}{
		{
			origin:   []byte{'+', '0', '5', '3', '2', '.', '0', '7'},
			intout:   532,
			floatout: 532.07,
		},
		{
			origin:   []byte{'+', '5', '3', '2', '.', '3', '7'},
			intout:   532,
			floatout: 532.37,
		},
		{
			origin:   []byte{'-', '5', '3', '2', '.', '3', '7'},
			intout:   -532,
			floatout: -532.37,
		},
		{
			origin:   []uint8{'-', '5', '3', '2', '.', '3', '7'},
			intout:   -532,
			floatout: -532.37,
		},
		{
			origin:   []uint8{'5', '3', '2', '.', '3', '7'},
			intout:   532,
			floatout: 532.37,
		},
		{
			origin:   10,
			intout:   10,
			floatout: 10.0,
		},
		{
			origin:   4.5,
			intout:   4,
			floatout: 4.5,
		},
		{
			origin:   []uint8{'5', '3', '2'},
			intout:   532,
			floatout: 532.0,
		},
	}
	ass := assert.New(t)
	for idx, tc := range testData {
		rr := resultResolve{tc.origin}
		ass.Equal(tc.intout, rr.Int64(), "case#%d fail", idx)
		ass.True(math.Abs(tc.floatout-rr.Float64()) < 1e-5, "case#%d fail", idx)
	}
}

func TestAggregateQuery(t *testing.T) {
	db, mock, err := sqlmock.New()
	if nil != err {
		t.Error(err)
	}
	var testData = []struct {
		rows     *sqlmock.Rows
		reg      string
		ag       AggregateSymbleBuilder
		intout   int64
		floatout float64
		where    map[string]any
	}{
		{
			sqlmock.NewRows([]string{"count(*)"}).AddRow(12),
			"SELECT count\\(\\*\\) FROM tb1",
			AggregateCount("*"),
			12,
			12,
			nil,
		},
		{
			sqlmock.NewRows([]string{"sum(age)"}).AddRow(math.MaxInt64),
			"SELECT sum\\(age\\) FROM tb1",
			AggregateSum("age"),
			math.MaxInt64,
			float64(math.MaxInt64),
			nil,
		},
		{
			sqlmock.NewRows([]string{"avg(age)"}).AddRow(100.957),
			"SELECT avg\\(age\\) FROM tb1",
			AggregateAvg("age"),
			100,
			100.957,
			nil,
		},
		{
			sqlmock.NewRows([]string{"sum(age)"}).AddRow(100.957),
			"sum\\(age\\)",
			AggregateSum("age"),
			100,
			100.957,
			nil,
		},
		{
			sqlmock.NewRows([]string{"max(age)"}).AddRow(math.Pi),
			"max\\(age\\)",
			AggregateMax("age"),
			3,
			math.Pi,
			nil,
		},
		{
			sqlmock.NewRows([]string{"min(age)"}).AddRow(math.Pi),
			"min\\(age\\) .+foo",
			AggregateMin("age"),
			3,
			math.Pi,
			map[string]any{"foo": "hello", "bar": 123},
		},
	}
	ass := assert.New(t)
	ctx := context.Background()
	for _, tc := range testData {
		mock.ExpectQuery(tc.reg).WillReturnRows(tc.rows)
		result, err := AggregateQuery(ctx, db, "tb1", tc.where, tc.ag)
		ass.NoError(err)
		ass.NoError(mock.ExpectationsWereMet())
		ass.Equal(tc.intout, result.Int64())
		ass.True(math.Abs(result.Float64()-tc.floatout) < 1e6)
	}
}

func TestOmitEmpty(t *testing.T) {
	var (
		m  map[string]string
		sl []string
		i  any
	)

	type stru struct {
		x string
		y int
	}

	var testData = []struct {
		where      map[string]any
		omitKey    []string
		finalWhere map[string]any
	}{
		// Bool
		{
			map[string]any{"b1": true, "b2": false},
			[]string{"b1", "b2", "b2"},
			map[string]any{"b1": true},
		},
		// Array, String
		{
			map[string]any{"a1": [0]string{}, "a2": [...]string{"2"}},
			[]string{"a1", "a2"},
			map[string]any{"a2": [...]string{"2"}},
		},
		// Float32, Float64
		{
			map[string]any{"f1": float32(0), "f2": float32(1.1)},
			[]string{"f1", "f2"},
			map[string]any{"f2": float32(1.1)},
		},
		{
			map[string]any{"f1": float64(0), "f2": float64(1.1)},
			[]string{"f1", "f2"},
			map[string]any{"f2": float64(1.1)},
		},
		// Int, Int8, Int16, Int32, Int64
		{
			map[string]any{"i8": int8(0), "i8_1": int8(8)},
			[]string{"i8", "i8_1"},
			map[string]any{"i8_1": int8(8)},
		},
		{
			map[string]any{"i16": int16(0), "i16_1": int16(16)},
			[]string{"i16", "i16_1"},
			map[string]any{"i16_1": int16(16)},
		},
		{
			map[string]any{"i32": int32(0), "i32_1": int32(32)},
			[]string{"i32", "i32_1"},
			map[string]any{"i32_1": int32(32)},
		},
		{
			map[string]any{"i64": int64(0), "i64_1": int64(64)},
			[]string{"i64", "i64_1"},
			map[string]any{"i64_1": int64(64)},
		},
		// Uint, Uint8, Uint16, Uint32, Uint64, Uintptr
		{
			map[string]any{"ui": uint(0), "ui_1": uint(1)},
			[]string{"ui", "ui_1"},
			map[string]any{"ui_1": uint(1)},
		},
		{
			map[string]any{"ui8": uint8(0), "ui8_1": uint8(8)},
			[]string{"ui8", "ui8_1"},
			map[string]any{"ui8_1": uint8(8)},
		},
		{
			map[string]any{"ui16": uint16(0), "ui16_1": uint16(16)},
			[]string{"ui16", "ui16_1"},
			map[string]any{"ui16_1": uint16(16)},
		},
		{
			map[string]any{"ui32": uint32(0), "ui32_1": uint32(32)},
			[]string{"ui32", "ui32_1"},
			map[string]any{"ui32_1": uint32(32)},
		},
		{
			map[string]any{"ui64": uint64(0), "ui64_1": uint64(64)},
			[]string{"ui64", "ui64_1"},
			map[string]any{"ui64_1": uint64(64)},
		},
		{
			map[string]any{"uip": uintptr(0), "uip_1": uintptr(1)},
			[]string{"uip", "uip_1"},
			map[string]any{"uip_1": uintptr(1)},
		},
		// Map, Slice, Interface
		{
			map[string]any{"m1": m, "m2": map[string]string{"foo": "hi"}, "m3": map[string]string{}},
			[]string{"m1", "m2", "m3"},
			map[string]any{"m2": map[string]string{"foo": "hi"}},
		},
		{
			map[string]any{"sl1": sl, "sl2": []string{"sl"}, "sl3": []int{}},
			[]string{"sl1", "sl2", "sl3"},
			map[string]any{"sl2": []string{"sl"}},
		},
		{
			map[string]any{"i": i},
			[]string{"i"},
			map[string]any{},
		},
		// struct
		{
			map[string]any{"stru1": stru{x: "s", y: 0}, "stru2": stru{x: "", y: 0}},
			[]string{"stru1", "stru2"},
			map[string]any{"stru1": stru{x: "s", y: 0}},
		},
		// implement zero
		{
			map[string]any{"time": time.Time{}},
			[]string{"time"},
			map[string]any{},
		},
	}
	ass := assert.New(t)
	for _, tc := range testData {
		r := OmitEmpty(tc.where, tc.omitKey)
		ass.True(reflect.DeepEqual(tc.finalWhere, r))
	}
}
