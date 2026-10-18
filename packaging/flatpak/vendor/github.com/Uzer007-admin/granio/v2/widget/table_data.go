package widget

import (
	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/data/binding"
)

// NewTableWithData displays a list of rows, each containing up to columns values.
// It also accepts binding.AsyncList[[]T]; background loads notify the table using
// the normal binding mechanism. Missing cells receive T's zero value.
func NewTableWithData[T any](data binding.List[[]T], columns int, create func() granio.CanvasObject, update func(T, granio.CanvasObject)) *Table {
	table := NewTable(func() (int, int) { return data.Length(), max(0, columns) }, create,
		func(id TableCellID, object granio.CanvasObject) {
			row, err := data.GetValue(id.Row)
			var value T
			if err == nil && id.Col < len(row) {
				value = row[id.Col]
			}
			update(value, object)
		})
	data.AddListener(binding.NewDataListener(table.Refresh))
	return table
}
