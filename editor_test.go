package terraedit

import (
	"math/rand"
	"testing"
)

func TestApplyReturnsOk(t *testing.T) {
	editor := NewEditor(map[string]string{"b": "builder"})
	if ok, _ := editor.Apply("b", 1, 1, 5); !ok {
		t.Fatal("有权限的编辑应该成功")
	}
}

func TestValueAfterApply(t *testing.T) {
	editor := NewEditor(map[string]string{"b": "builder"})
	editor.Apply("b", 2, 3, 7)
	if editor.Value(2, 3) != 7 {
		t.Fatalf("地块值不对：%d", editor.Value(2, 3))
	}
}

func TestUnknownCellIsZero(t *testing.T) {
	if NewEditor(nil).Value(9, 9) != 0 {
		t.Fatal("没编辑过的地块应该是 0")
	}
}

func TestScannedStartsAtZero(t *testing.T) {
	if NewEditor(nil).Scanned() != 0 {
		t.Fatal("初始扫描数应该是 0")
	}
}

func TestPermissionDenied(t *testing.T) {
	editor := NewEditor(map[string]string{"g": "guest", "v": "viewer"})
	for _, p := range []string{"g", "v", "nobody"} {
		if ok, reason := editor.Apply(p, 1, 1, 5); ok || reason != "denied" {
			t.Fatalf("%s 应该被拒绝，得到 (%v, %q)", p, ok, reason)
		}
	}
	if editor.Value(1, 1) != 0 {
		t.Fatal("被拒绝的编辑不应该改状态")
	}
	if editor.undoDepth() != 0 {
		t.Fatal("被拒绝的编辑不应该入撤销栈")
	}
	editor.Apply("g", 1, 1, 9)
	if editor.Undo() {
		t.Fatal("没有有效编辑时撤销应该返回 false")
	}
}

func TestAdminAndBuilderCanEdit(t *testing.T) {
	editor := NewEditor(map[string]string{"a": "admin", "b": "builder"})
	if ok, reason := editor.Apply("a", 0, 0, 1); !ok || reason != "ok" {
		t.Fatalf("admin 应该能编辑，得到 (%v, %q)", ok, reason)
	}
	if ok, reason := editor.Apply("b", 0, 0, 2); !ok || reason != "ok" {
		t.Fatalf("builder 应该能编辑，得到 (%v, %q)", ok, reason)
	}
}

func TestDuplicateEdit(t *testing.T) {
	editor := NewEditor(map[string]string{"b": "builder"})
	if ok, _ := editor.Apply("b", 1, 1, 5); !ok {
		t.Fatal("首次编辑应该成功")
	}
	if ok, reason := editor.Apply("b", 1, 1, 5); ok || reason != "duplicate" {
		t.Fatalf("重复设置当前值应返回 (false, duplicate)，得到 (%v, %q)", ok, reason)
	}
	if editor.undoDepth() != 1 {
		t.Fatalf("重复编辑不应入撤销栈，深度应为 1，得到 %d", editor.undoDepth())
	}
	if editor.Value(1, 1) != 5 {
		t.Fatal("重复编辑不应改地块值")
	}
	// 未编辑过的格子设成 0（当前值）也是重复。
	if ok, reason := editor.Apply("b", 9, 9, 0); ok || reason != "duplicate" {
		t.Fatalf("新格设 0 应判重，得到 (%v, %q)", ok, reason)
	}
}

func TestUndoRestoresPreviousValue(t *testing.T) {
	editor := NewEditor(map[string]string{"b": "builder"})
	editor.Apply("b", 1, 1, 5)
	editor.Apply("b", 1, 1, 8)
	editor.Apply("b", 2, 2, 3)
	if !editor.Undo() || editor.Value(2, 2) != 0 {
		t.Fatal("第一次撤销应恢复 (2,2) 编辑前的 0")
	}
	if !editor.Undo() || editor.Value(1, 1) != 5 {
		t.Fatal("第二次撤销应恢复 (1,1) 编辑前的 5")
	}
	if !editor.Undo() || editor.Value(1, 1) != 0 {
		t.Fatal("第三次撤销应恢复 (1,1) 最初的 0")
	}
	if editor.Undo() {
		t.Fatal("栈空时撤销应返回 false")
	}
}

func TestScannedNotLinearInCellCount(t *testing.T) {
	editor := NewEditor(map[string]string{"b": "builder"})
	for i := 0; i < 1000; i++ {
		editor.Apply("b", i, i, i)
	}
	before := editor.Scanned()
	editor.Value(0, 0)
	editor.Value(500, 500)
	if got := editor.Scanned() - before; got != 2 {
		t.Fatalf("两次查询应只扫 2 格，得到 %d", got)
	}
}

func TestDeterministicReplay(t *testing.T) {
	ops := []struct {
		player      string
		x, y, value int
		undo        bool
	}{
		{"g", 0, 0, 5, false},
		{"b", 0, 0, 5, false},
		{"b", 0, 0, 5, false},
		{"b", 1, 1, 7, false},
		{"b", 1, 1, 4, false},
		{"", 0, 0, 0, true},
		{"b", 1, 1, 7, false},
		{"", 0, 0, 0, true},
		{"", 0, 0, 0, true},
		{"", 0, 0, 0, true},
	}
	run := func() (string, int) {
		editor := NewEditor(map[string]string{"g": "guest", "b": "builder", "a": "admin"})
		var seq []byte
		for _, op := range ops {
			if op.undo {
				if editor.Undo() {
					seq = append(seq, '1')
				} else {
					seq = append(seq, '0')
				}
			} else {
				ok, reason := editor.Apply(op.player, op.x, op.y, op.value)
				if ok {
					seq = append(seq, 'o')
				} else {
					seq = append(seq, reason[0])
				}
			}
		}
		return string(seq), editor.undoDepth()
	}
	first, depth := run()
	second, _ := run()
	if first != second || depth != 0 {
		t.Fatalf("同一序列重复执行结果应相同且栈清空，得到 %q vs %q，深度 %d", first, second, depth)
	}
}

// refEditor 是「权限校验 + 编辑幂等 + 撤销栈」的独立参考实现。
type refEditor struct {
	roles map[string]string
	cells map[[2]int]int
	stack []refEdit
}

type refEdit struct {
	x, y, old int
}

func newRef(roles map[string]string) *refEditor {
	return &refEditor{roles: roles, cells: map[[2]int]int{}}
}

func (r *refEditor) apply(player string, x, y, value int) (bool, string) {
	role, ok := r.roles[player]
	if !ok || (role != "admin" && role != "builder") {
		return false, "denied"
	}
	key := [2]int{x, y}
	if r.cells[key] == value {
		return false, "duplicate"
	}
	r.stack = append(r.stack, refEdit{x: x, y: y, old: r.cells[key]})
	r.cells[key] = value
	return true, "ok"
}

func (r *refEditor) undo() bool {
	if len(r.stack) == 0 {
		return false
	}
	last := r.stack[len(r.stack)-1]
	r.stack = r.stack[:len(r.stack)-1]
	r.cells[[2]int{last.x, last.y}] = last.old
	return true
}

func (e *Editor) undoDepth() int { return len(e.stack) }

func TestDifferentialAgainstReference(t *testing.T) {
	roles := map[string]string{"a": "admin", "b": "builder", "g": "guest"}
	players := []string{"a", "b", "g", "stranger"}
	const sequences = 200
	for seed := int64(0); seed < sequences; seed++ {
		rng := rand.New(rand.NewSource(seed))
		got := NewEditor(roles)
		want := newRef(roles)
		for step := 0; step < 40; step++ {
			switch rng.Intn(6) {
			case 5:
				if got.Undo() != want.undo() {
					t.Fatalf("seed=%d step=%d Undo 结果不一致", seed, step)
				}
			default:
				p := players[rng.Intn(len(players))]
				x, y := rng.Intn(5), rng.Intn(5)
				value := rng.Intn(4)
				gok, greason := got.Apply(p, x, y, value)
				wok, wreason := want.apply(p, x, y, value)
				if gok != wok || greason != wreason {
					t.Fatalf("seed=%d step=%d Apply(%s,%d,%d,%d) 结果不一致：(%v,%q) vs (%v,%q)",
						seed, step, p, x, y, value, gok, greason, wok, wreason)
				}
			}
			for cx := 0; cx < 5; cx++ {
				for cy := 0; cy < 5; cy++ {
					if got.Value(cx, cy) != want.cells[[2]int{cx, cy}] {
						t.Fatalf("seed=%d step=%d 地块 (%d,%d) 值不一致", seed, step, cx, cy)
					}
				}
			}
			if got.undoDepth() != len(want.stack) {
				t.Fatalf("seed=%d step=%d 撤销栈深度不一致：%d vs %d",
					seed, step, got.undoDepth(), len(want.stack))
			}
		}
	}
}
