// Package terraedit 是地块编辑：权限校验、编辑幂等、撤销栈、按地块索引。
// 缺陷：不校验权限、重复编辑也入撤销栈、撤销没有实现、每次查询全扫地块表。
package terraedit

// Editor 是地块编辑器。
type Editor struct {
	roles   map[string]string
	cells   map[[2]int]int
	stack   [][2]int
	scanned int
}

// NewEditor 按角色表建编辑器。
func NewEditor(roles map[string]string) *Editor {
	return &Editor{roles: roles, cells: map[[2]int]int{}}
}

// Apply 修改一个地块。缺陷：不校验角色权限、重复编辑也压栈。
func (e *Editor) Apply(player string, x, y, value int) (bool, string) {
	e.cells[[2]int{x, y}] = value
	e.stack = append(e.stack, [2]int{x, y})
	return true, "ok"
}

// Undo 撤销上一次编辑。缺陷：没有实现。
func (e *Editor) Undo() bool { return false }

// Value 读一个地块。缺陷：每次全扫地块表。
func (e *Editor) Value(x, y int) int {
	e.scanned += len(e.cells)
	return e.cells[[2]int{x, y}]
}

// Scanned 是累计扫描的地块数（规模观测）。
func (e *Editor) Scanned() int { return e.scanned }
