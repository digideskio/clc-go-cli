package formatters

import (
	"bytes"
	"fmt"
	table "github.com/ldmberman/tablewriter"
	"strings"
)

type TableFormatter struct{}

func (f *TableFormatter) FormatOutput(model interface{}) (res string, err error) {
	return getTable(model, 0)
}

// getTable constructs a textual table representation of the input model, which
// should be of either []interface{} or map[string]interface{} type.
//
// Elements of a slice are rendered one after another separated by a newline.
//
// Maps' values that are themselves maps or slices are rendered as tables, which
// are, in turn, placed each in a separate table row. The rest of the values are
// rendered as a table occupying the first row of the outer table.
//
// Elements that are neither maps nor slices are rendered in Go %v format except for
// the nil values: cell content for them is empty.
//
// Every table can either have map keys in the first row and map values in the second
// or map keys in the first column and map values in the second. The latter is chosen
// over the former if the terminal width is not enough to display all the keys and/or
// values in one row.
func getTable(m interface{}, depth int) (string, error) {
	if mmap, ok := m.(map[string]interface{}); ok {
		if len(mmap) == 0 {
			return "", nil
		}

		// keys and values store map entries whose values are neither maps nor slices.
		// nested stores maps and slices found in the map.
		keys, values, nested := []string{}, []string{}, map[string]interface{}{}
		for _, k := range sortedKeys(mmap) {
			value, err := getTable(mmap[k], depth+1)
			if err != nil {
				return "", err
			}
			_, isMap := mmap[k].(map[string]interface{})
			_, isSlice := mmap[k].([]interface{})
			if isMap || isSlice {
				nested[k] = value
			} else {
				keys = append(keys, k)
				values = append(values, value)
			}
		}
		if len(nested) == 0 {
			return innerTable(keys, values, depth, true, false), nil
		}
		// outerT contains a table with non-map,non-slice values in the first
		// row and tables with other values in the subsequent rows.
		outerT, outerBuf := setupTable(false)
		if len(keys) > 0 {
			outerT.Append([]string{innerTable(keys, values, depth+1, true, false)})
		}
		for _, k := range sortedKeys(nested) {
			outerT.Append([]string{innerTable([]string{k}, []string{nested[k].(string)}, depth+1, false, true)})
		}
		outerT.Render()
		return outerBuf.String(), nil
	} else if mslice, ok := m.([]interface{}); ok {
		// Elements of a slice are simply returned one after another
		// separated by a newline.
		values := []string{}
		for _, el := range mslice {
			v, err := getTable(el, depth)
			if err != nil {
				return "", err
			}
			values = append(values, fmt.Sprintf("%v", v))
		}
		return strings.Join(values, "\n"), nil
	} else {
		if m == nil {
			return "", nil
		}
		return fmt.Sprintf("%v", m), nil
	}
}

func setupTable(autowrap bool) (*table.Table, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	t := table.NewWriter(buf)
	t.SetRowLine(true)
	if autowrap {
		t.SetAlignment(table.ALIGN_CENTRE)
	} else {
		t.SetAlignment(table.ALIGN_LEFT)
	}
	t.SetAutoWrapText(autowrap)
	return t, buf
}

func innerTable(keys, values []string, depth int, autowrap bool, vertical bool) string {
	t, buf := setupTable(autowrap)
	if vertical || getMinWidth(keys, values, depth) < getTerminalWidth() {
		t.Append(keys)
		t.Append(values)
		t.Render()
	} else {
		for i, k := range keys {
			t.Append([]string{k, values[i]})
		}
		t.Render()
	}
	return buf.String()
}

func getMinWidth(keys, values []string, depth int) uint {
	delimiters := (len(keys) + 1) + depth*2

	// 2 spaces around each of the cell values
	// + 2 spaces around outer tables' borders for each level of nesting
	paddings := len(keys)*2 + depth*2

	words := 0
	for i, k := range keys {
		if len(k) > len(values[i]) {
			words += len(k)
		} else {
			words += len(values[i])
		}
	}
	return uint(words + paddings + delimiters)
}
