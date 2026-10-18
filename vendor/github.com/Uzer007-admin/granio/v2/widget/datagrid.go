package widget

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Uzer007-admin/granio/v2"
	"github.com/Uzer007-admin/granio/v2/layout"
)

// DataGridSource предоставляет строки с устойчивыми уникальными идентификаторами.
// Все методы вызываются в потоке интерфейса; изменение данных завершается Reload.
type DataGridSource interface {
	RowIDs() []string
	Value(rowID, columnID string) string
}

// EditableDataGridSource добавляет транзакционную запись значения ячейки.
type EditableDataGridSource interface {
	DataGridSource
	SetValue(rowID, columnID, value string) error
}

// DataGridColumn задаёт представление, сравнение и проверку данных колонки.
type DataGridColumn struct {
	ID, Title string
	Width     float32
	Editable  bool
	Compare   func(a, b string) int
	Validate  func(string) error
}

// DataGridCell сохраняет идентичность ячейки при сортировке и фильтрации.
type DataGridCell struct{ RowID, ColumnID string }

type dataGridRow struct {
	id, group string
	count     int
	header    bool
}
type dataGridEdit struct {
	cell       DataGridCell
	draft      string
	start, end int
}

// DataGrid добавляет проекцию данных и редактирование к виртуализированной Table.
// Table предназначена для настройки ширины, закрепления и представления;
// её callbacks выбора обслуживаются DataGrid.
type DataGrid struct {
	BaseWidget
	Table                   *Table
	OnSelected              func(DataGridCell)
	OnEdited                func(DataGridCell, string)
	OnEditError             func(error)
	columns                 []DataGridColumn
	source                  DataGridSource
	rows                    []dataGridRow
	sourceIDs               map[string]bool
	rowIndex                map[string]int
	columnIndex             map[string]int
	selection               map[DataGridCell]bool
	filter                  func(string) bool
	sortColumn, groupColumn string
	descending, restoring   bool
	edit                    *dataGridEdit
	activeEditor            *dataGridEntry
}

var ErrDataGridCell = errors.New("ячейка отсутствует в текущем представлении")
var ErrDataGridReadOnly = errors.New("редактирование ячейки недоступно")

func NewDataGrid(columns []DataGridColumn, source DataGridSource) (*DataGrid, error) {
	if source == nil || len(columns) == 0 {
		return nil, errors.New("нужны источник данных и колонки")
	}
	g := &DataGrid{source: source, columns: append([]DataGridColumn(nil), columns...), columnIndex: make(map[string]int, len(columns)), selection: make(map[DataGridCell]bool)}
	for index, c := range columns {
		if c.ID == "" || c.Width < 0 || math.IsNaN(float64(c.Width)) || math.IsInf(float64(c.Width), 0) {
			return nil, errors.New("некорректное описание колонки")
		}
		if _, ok := g.columnIndex[c.ID]; ok {
			return nil, fmt.Errorf("повторная колонка: %s", c.ID)
		}
		g.columnIndex[c.ID] = index
	}
	g.ExtendBaseWidget(g)
	g.Table = NewTable(func() (int, int) { return len(g.rows), len(g.columns) }, func() granio.CanvasObject { return newDataGridCell(g) }, g.updateCell)
	g.Table.SelectionMode = SelectionMultiple
	g.Table.ShowHeaderRow = true
	g.Table.CreateHeader = func() granio.CanvasObject { b := NewButton("", nil); b.Importance = LowImportance; return b }
	g.Table.UpdateHeader = func(id TableCellID, object granio.CanvasObject) {
		b := object.(*Button)
		if id.Col < 0 || id.Col >= len(g.columns) {
			b.SetText("")
			b.OnTapped = nil
			return
		}
		c := g.columns[id.Col]
		title := c.Title
		if c.ID == g.sortColumn {
			if g.descending {
				title += " ↓"
			} else {
				title += " ↑"
			}
		}
		b.SetText(title)
		b.OnTapped = func() {
			if g.sortColumn == c.ID && g.descending {
				_ = g.SetSort("", false)
			} else {
				_ = g.SetSort(c.ID, g.sortColumn == c.ID)
			}
		}
	}
	g.Table.OnSelected = func(id TableCellID) {
		if g.restoring {
			return
		}
		cell, ok := g.cellAt(id)
		if !ok {
			g.restoreSelection()
			return
		}
		g.selection[cell] = true
		if g.OnSelected != nil {
			g.OnSelected(cell)
		}
	}
	g.Table.OnUnselected = func(id TableCellID) {
		if g.restoring {
			return
		}
		if cell, ok := g.cellAt(id); ok {
			delete(g.selection, cell)
		}
	}
	g.Table.OnEdit = func(id TableCellID) {
		if cell, ok := g.cellAt(id); ok {
			if err := g.BeginEdit(cell); err != nil && g.OnEditError != nil {
				g.OnEditError(err)
			}
		}
	}
	for i, c := range columns {
		if c.Width > 0 {
			g.Table.SetColumnWidth(i, c.Width)
		}
	}
	if err := g.Reload(); err != nil {
		return nil, err
	}
	return g, nil
}

func (g *DataGrid) CreateRenderer() granio.WidgetRenderer { return NewSimpleRenderer(g.Table) }

func (g *DataGrid) cellAt(id TableCellID) (DataGridCell, bool) {
	if id.Row < 0 || id.Row >= len(g.rows) || id.Col < 0 || id.Col >= len(g.columns) || g.rows[id.Row].header {
		return DataGridCell{}, false
	}
	return DataGridCell{g.rows[id.Row].id, g.columns[id.Col].ID}, true
}

// ViewRow возвращает идентификатор строки либо заголовок группы.
func (g *DataGrid) ViewRow(index int) (rowID, group string, header bool) {
	if index < 0 || index >= len(g.rows) {
		return "", "", false
	}
	row := g.rows[index]
	return row.id, row.group, row.header
}
func (g *DataGrid) ViewRowCount() int { return len(g.rows) }

// Reload строит проекцию до изменения представления: ошибка источника сохраняет прежний вид.
func (g *DataGrid) Reload() error {
	ids := append([]string(nil), g.source.RowIDs()...)
	seen := make(map[string]bool, len(ids))
	type item struct{ id, sortKey, groupKey string }
	items := make([]item, 0, len(ids))
	for _, id := range ids {
		if id == "" || seen[id] {
			return errors.New("идентификаторы строк должны быть уникальными и непустыми")
		}
		seen[id] = true
		if g.filter != nil && !g.filter(id) {
			continue
		}
		row := item{id: id}
		if g.sortColumn != "" {
			row.sortKey = g.source.Value(id, g.sortColumn)
		}
		if g.groupColumn != "" {
			row.groupKey = g.source.Value(id, g.groupColumn)
		}
		items = append(items, row)
	}
	if g.sortColumn != "" {
		compare := g.columns[g.columnIndex[g.sortColumn]].Compare
		if compare == nil {
			compare = strings.Compare
		}
		sort.SliceStable(items, func(i, j int) bool {
			order := compare(items[i].sortKey, items[j].sortKey)
			if g.descending {
				return order > 0
			}
			return order < 0
		})
	}
	rows := make([]dataGridRow, 0, len(items))
	if g.groupColumn == "" {
		for _, row := range items {
			rows = append(rows, dataGridRow{id: row.id})
		}
	} else {
		grouped := make(map[string][]string)
		for _, row := range items {
			grouped[row.groupKey] = append(grouped[row.groupKey], row.id)
		}
		keys := make([]string, 0, len(grouped))
		for key := range grouped {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			rows = append(rows, dataGridRow{group: key, count: len(grouped[key]), header: true})
			for _, id := range grouped[key] {
				rows = append(rows, dataGridRow{id: id})
			}
		}
	}
	if g.edit != nil && g.activeEditor != nil {
		_, g.edit.start, g.edit.end = g.activeEditor.AccessibilityTextSelection()
	}
	g.rows, g.sourceIDs = rows, seen
	g.rowIndex = make(map[string]int, len(items))
	for index, row := range rows {
		if !row.header {
			g.rowIndex[row.id] = index
		}
	}
	for cell := range g.selection {
		if !seen[cell.RowID] {
			delete(g.selection, cell)
		}
	}
	if g.edit != nil && !seen[g.edit.cell.RowID] {
		g.edit = nil
	}
	previousEditor := g.activeEditor
	hadFocus := false
	if app := granio.CurrentApp(); app != nil {
		if c := app.Driver().CanvasForObject(g); c != nil {
			hadFocus = previousEditor != nil && c.Focused() == previousEditor
		}
	}
	g.activeEditor = nil
	g.restoreSelection()
	g.Table.Refresh()
	if g.edit != nil && g.activeEditor != nil && g.activeEditor != previousEditor {
		g.activeEditor.AccessibilitySetSelection(g.edit.start, g.edit.end)
	}
	if hadFocus && g.activeEditor != previousEditor {
		if g.activeEditor != nil {
			g.focus(g.activeEditor)
		} else {
			g.focus(g.Table)
		}
	}
	return nil
}

func (g *DataGrid) SetFilter(filter func(rowID string) bool) error {
	previous := g.filter
	g.filter = filter
	if err := g.Reload(); err != nil {
		g.filter = previous
		return err
	}
	return nil
}
func (g *DataGrid) SetSort(columnID string, descending bool) error {
	if _, ok := g.columnIndex[columnID]; columnID != "" && !ok {
		return ErrDataGridCell
	}
	previous, previousDirection := g.sortColumn, g.descending
	g.sortColumn, g.descending = columnID, descending
	if err := g.Reload(); err != nil {
		g.sortColumn, g.descending = previous, previousDirection
		return err
	}
	return nil
}
func (g *DataGrid) SetGroupColumn(columnID string) error {
	if _, ok := g.columnIndex[columnID]; columnID != "" && !ok {
		return ErrDataGridCell
	}
	previous := g.groupColumn
	g.groupColumn = columnID
	if err := g.Reload(); err != nil {
		g.groupColumn = previous
		return err
	}
	return nil
}

// SelectedCells включает временно скрытые фильтром ячейки.
func (g *DataGrid) SelectedCells() []DataGridCell {
	cells := make([]DataGridCell, 0, len(g.selection))
	for cell := range g.selection {
		cells = append(cells, cell)
	}
	sort.Slice(cells, func(i, j int) bool {
		if cells[i].RowID == cells[j].RowID {
			return cells[i].ColumnID < cells[j].ColumnID
		}
		return cells[i].RowID < cells[j].RowID
	})
	return cells
}
func (g *DataGrid) SetSelection(cells []DataGridCell) {
	next := make(map[DataGridCell]bool, len(cells))
	for _, cell := range cells {
		if _, ok := g.columnIndex[cell.ColumnID]; ok && g.sourceIDs[cell.RowID] {
			next[cell] = true
		}
	}
	g.selection = next
	g.restoreSelection()
}
func (g *DataGrid) restoreSelection() {
	selected := make([]TableCellID, 0, len(g.selection))
	for cell := range g.selection {
		if row, ok := g.rowIndex[cell.RowID]; ok {
			selected = append(selected, TableCellID{row, g.columnIndex[cell.ColumnID]})
		}
	}
	sortCells(selected)
	g.restoring = true
	defer func() { g.restoring = false }()
	g.Table.SetSelection(selected)
}

func (g *DataGrid) BeginEdit(cell DataGridCell) error {
	column, ok := g.columnIndex[cell.ColumnID]
	if !ok {
		return ErrDataGridCell
	}
	row, visible := g.rowIndex[cell.RowID]
	if !visible {
		return ErrDataGridCell
	}
	if _, ok := g.source.(EditableDataGridSource); !ok || !g.columns[column].Editable {
		return ErrDataGridReadOnly
	}
	value := g.source.Value(cell.RowID, cell.ColumnID)
	g.activeEditor = nil
	g.edit = &dataGridEdit{cell: cell, draft: value, start: 0, end: len([]rune(value))}
	g.Table.ScrollTo(TableCellID{row, column})
	g.Table.Refresh()
	if g.activeEditor != nil {
		g.activeEditor.AccessibilitySetSelection(0, len([]rune(value)))
		g.focus(g.activeEditor)
	}
	return nil
}
func (g *DataGrid) EditingCell() (DataGridCell, string, bool) {
	if g.edit == nil {
		return DataGridCell{}, "", false
	}
	return g.edit.cell, g.edit.draft, true
}
func (g *DataGrid) CommitEdit(value string) error {
	if g.edit == nil {
		return ErrDataGridCell
	}
	cell := g.edit.cell
	column := g.columns[g.columnIndex[cell.ColumnID]]
	g.edit.draft = value
	if g.activeEditor != nil && g.activeEditor.Text != value {
		g.activeEditor.SetText(value)
	}
	var err error
	if column.Validate != nil {
		err = column.Validate(value)
	}
	if err == nil {
		source, ok := g.source.(EditableDataGridSource)
		if !ok || !column.Editable || !g.sourceIDs[cell.RowID] {
			err = ErrDataGridReadOnly
		} else {
			err = source.SetValue(cell.RowID, cell.ColumnID, value)
		}
	}
	if err != nil {
		if g.activeEditor != nil {
			g.activeEditor.SetValidationError(err)
		}
		if g.OnEditError != nil {
			g.OnEditError(err)
		}
		return err
	}
	g.edit = nil
	g.activeEditor = nil
	if err = g.Reload(); err != nil {
		return err
	}
	g.focus(g.Table)
	if g.OnEdited != nil {
		g.OnEdited(cell, value)
	}
	return nil
}
func (g *DataGrid) CancelEdit() {
	if g.edit == nil {
		return
	}
	g.edit = nil
	g.activeEditor = nil
	g.Table.Refresh()
	g.focus(g.Table)
}
func (g *DataGrid) focus(object granio.Focusable) {
	if app := granio.CurrentApp(); app != nil {
		if c := app.Driver().CanvasForObject(g); c != nil {
			c.Focus(object)
		}
	}
}

type dataGridCell struct {
	BaseWidget
	grid   *DataGrid
	id     DataGridCell
	label  *Label
	editor *dataGridEntry
}

func newDataGridCell(g *DataGrid) *dataGridCell {
	c := &dataGridCell{grid: g, label: NewLabel("")}
	c.ExtendBaseWidget(c)
	c.editor = &dataGridEntry{grid: g}
	c.editor.ExtendBaseWidget(c.editor)
	c.editor.Hide()
	c.editor.OnChanged = func(value string) {
		if g.edit != nil && g.edit.cell == c.id {
			g.edit.draft = value
		}
	}
	c.editor.OnSubmitted = func(value string) {
		if g.edit != nil && g.edit.cell == c.id && g.activeEditor == c.editor {
			_ = g.CommitEdit(value)
		}
	}
	return c
}
func (c *dataGridCell) CreateRenderer() granio.WidgetRenderer {
	return NewSimpleRenderer(granio.NewContainerWithLayout(layout.NewStackLayout(), c.label, c.editor))
}
func (c *dataGridCell) MinSize() granio.Size            { return c.label.MinSize().Max(c.editor.MinSize()) }
func (c *dataGridCell) DoubleTapped(*granio.PointEvent) { _ = c.grid.BeginEdit(c.id) }
func (c *dataGridCell) releaseTableCell() {
	g := c.grid
	if g.activeEditor != c.editor {
		return
	}
	if g.edit != nil && g.edit.cell == c.id {
		_, g.edit.start, g.edit.end = c.editor.AccessibilityTextSelection()
	}
	g.activeEditor = nil
	if app := granio.CurrentApp(); app != nil {
		if surface := app.Driver().CanvasForObject(g); surface != nil && surface.Focused() == c.editor {
			surface.Focus(g.Table)
		}
	}
}
func (g *DataGrid) updateCell(id TableCellID, object granio.CanvasObject) {
	c := object.(*dataGridCell)
	row := g.rows[id.Row]
	column := g.columns[id.Col]
	next := DataGridCell{row.id, column.ID}
	if c.id != next {
		c.releaseTableCell()
	}
	c.id = next
	c.label.TextStyle.Bold = row.header
	if row.header {
		value := ""
		if id.Col == 0 {
			value = fmt.Sprintf("%s (%d)", row.group, row.count)
		}
		c.label.SetText(value)
	} else {
		c.label.SetText(g.source.Value(row.id, column.ID))
	}
	if g.edit != nil && !row.header && g.edit.cell == c.id {
		c.editor.Validator = column.Validate
		if c.editor.Text != g.edit.draft {
			c.editor.SetText(g.edit.draft)
		}
		c.label.Hide()
		c.editor.Show()
		if g.activeEditor != c.editor {
			g.activeEditor = c.editor
			c.editor.AccessibilitySetSelection(g.edit.start, g.edit.end)
		}
	} else {
		c.editor.Hide()
		c.label.Show()
	}
}

type dataGridEntry struct {
	Entry
	grid *DataGrid
}

func (e *dataGridEntry) TypedKey(event *granio.KeyEvent) {
	if event.Name == granio.KeyEscape {
		e.grid.CancelEdit()
		return
	}
	e.Entry.TypedKey(event)
}
