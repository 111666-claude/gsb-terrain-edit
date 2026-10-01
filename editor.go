// Package terraedit 是地块编辑：权限校验、编辑幂等、撤销栈、按地块索引。
package terraedit

// edit 记录一次编辑前的状态，用于撤销。
type edit struct {
	x, y     int
	oldValue int
}

// Editor 是地块编辑器。
type Editor struct {
	roles   map[string]string
	cells   map[[2]int]int
	stack   []edit
	scanned int
}

// NewEditor 按角色表建编辑器。
func NewEditor(roles map[string]string) *Editor {
	return &Editor{roles: roles, cells: map[[2]int]int{}}
}

// Apply 修改一个地块。只有 admin 或 builder 能编辑；设成当前值是重复编辑。
func (e *Editor) Apply(player string, x, y, value int) (bool, string) {
	if role := e.roles[player]; role != "admin" && role != "builder" {
		return false, "denied"
	}
	key := [2]int{x, y}
	old := e.cells[key]
	if old == value {
		return false, "duplicate"
	}
	e.cells[[2]int{x, y}] = value
	e.stack = append(e.stack, edit{x: x, y: y, oldValue: old})
	return true, "ok"
}

// Undo 按后进先出恢复上一次编辑前的值，栈空返回 false。
func (e *Editor) Undo() bool {
	if len(e.stack) == 0 {
		return false
	}
	last := e.stack[len(e.stack)-1]
	e.stack = e.stack[:len(e.stack)-1]
	e.cells[[2]int{last.x, last.y}] = last.oldValue
	return true
}

// Value 读一个地块，单次 O(1)。
func (e *Editor) Value(x, y int) int {
	e.scanned++
	return e.cells[[2]int{x, y}]
}

// Scanned 是累计扫描的地块数（规模观测）。
func (e *Editor) Scanned() int { return e.scanned }
