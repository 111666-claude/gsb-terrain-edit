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
	editor := NewEditor(map[string]string{"p": "guest", "b": "builder", "a": "admin"})
	if ok, reason := editor.Apply("p", 1, 1, 5); ok || reason != "denied" {
		t.Fatalf("guest 应该被拒绝，得到 (%v, %q)", ok, reason)
	}
	if ok, reason := editor.Apply("nobody", 1, 1, 5); ok || reason != "denied" {
		t.Fatalf("未知玩家应该被拒绝，得到 (%v, %q)", ok, reason)
	}
	if editor.Value(1, 1) != 0 || len(editor.stack) != 0 {
		t.Fatal("被拒绝的编辑不该改状态")
	}
	if ok, _ := editor.Apply("a", 1, 1, 5); !ok {
		t.Fatal("admin 应该能编辑")
	}
	if ok, _ := editor.Apply("b", 1, 1, 6); !ok {
		t.Fatal("builder 应该能编辑")
	}
}

func TestDuplicateEdit(t *testing.T) {
	editor := NewEditor(map[string]string{"b": "builder"})
	editor.Apply("b", 1, 1, 5)
	if ok, reason := editor.Apply("b", 1, 1, 5); ok || reason != "duplicate" {
		t.Fatalf("重复编辑应返回 (false, duplicate)，得到 (%v, %q)", ok, reason)
	}
	if len(editor.stack) != 1 {
		t.Fatalf("重复编辑不该入栈，栈深 %d", len(editor.stack))
	}
	if ok, reason := editor.Apply("b", 9, 9, 0); ok || reason != "duplicate" {
		t.Fatalf("未编辑地块设 0 也是重复，得到 (%v, %q)", ok, reason)
	}
	if len(editor.stack) != 1 {
		t.Fatal("对未编辑地块设 0 不该入栈")
	}
}

func TestUndoSemantics(t *testing.T) {
	editor := NewEditor(map[string]string{"b": "builder"})
	if editor.Undo() {
		t.Fatal("空栈撤销应返回 false")
	}
	editor.Apply("b", 1, 1, 5)
	editor.Apply("b", 1, 1, 9)
	editor.Apply("b", 2, 2, 3)
	if !editor.Undo() || editor.Value(2, 2) != 0 {
		t.Fatal("撤销应恢复最后一次编辑前的值")
	}
	if !editor.Undo() || editor.Value(1, 1) != 5 {
		t.Fatal("撤销应按后进先出恢复")
	}
	if !editor.Undo() || editor.Value(1, 1) != 0 {
		t.Fatal("撤销应恢复到初始值")
	}
	if editor.Undo() {
		t.Fatal("栈再次为空时撤销应返回 false")
	}
}

func TestUndoThenReeditIsNotDuplicate(t *testing.T) {
	editor := NewEditor(map[string]string{"b": "builder"})
	editor.Apply("b", 1, 1, 5)
	editor.Undo()
	if ok, reason := editor.Apply("b", 1, 1, 5); !ok || reason != "ok" {
		t.Fatalf("撤销后重新设同样的值不是重复，得到 (%v, %q)", ok, reason)
	}
	if !editor.Undo() || editor.Value(1, 1) != 0 {
		t.Fatal("撤销栈应在重新编辑后恢复工作")
	}
}

func TestScannedDoesNotGrowWithCells(t *testing.T) {
	editor := NewEditor(map[string]string{"b": "builder"})
	for i := 0; i < 1000; i++ {
		editor.Apply("b", i, i, i)
	}
	before := editor.Scanned()
	editor.Value(0, 0)
	editor.Value(1, 1)
	if got := editor.Scanned() - before; got != 2 {
		t.Fatalf("两次查询 Scanned 应只增 2，实际增 %d", got)
	}
}

type refEdit struct {
	x, y, old int
}

type refEditor struct {
	roles map[string]string
	cells map[[2]int]int
	stack []refEdit
}

func newRef(roles map[string]string) *refEditor {
	return &refEditor{roles: roles, cells: map[[2]int]int{}}
}

func (r *refEditor) apply(player string, x, y, value int) (bool, string) {
	if role := r.roles[player]; role != "admin" && role != "builder" {
		return false, "denied"
	}
	key := [2]int{x, y}
	if r.cells[key] == value {
		return false, "duplicate"
	}
	r.stack = append(r.stack, refEdit{x, y, r.cells[key]})
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

func (r *refEditor) value(x, y int) int { return r.cells[[2]int{x, y}] }

func TestRandomMatchesReference(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	roles := map[string]string{"a": "admin", "b": "builder", "g": "guest"}
	players := []string{"a", "b", "g", "ghost"}

	for seq := 0; seq < 200; seq++ {
		got := NewEditor(roles)
		want := newRef(roles)

		// 每条序列跑两遍，验证同一串操作重复执行结果相同。
		for run := 0; run < 2; run++ {
			if run == 1 {
				got = NewEditor(roles)
				want = newRef(roles)
			}
			n := rng.Intn(60)
			for i := 0; i < n; i++ {
				player := players[rng.Intn(len(players))]
				x, y := rng.Intn(8), rng.Intn(8)
				switch rng.Intn(3) {
				case 0, 1: // 编辑；约一半概率设成当前值制造重复
					value := rng.Intn(10)
					if rng.Intn(2) == 0 {
						value = got.Value(x, y)
					}
					gok, greason := got.Apply(player, x, y, value)
					wok, wreason := want.apply(player, x, y, value)
					if gok != wok || greason != wreason {
						t.Fatalf("seq %d run %d Apply(%q,%d,%d,%d) 结果不符：(%v,%q) vs (%v,%q)",
							seq, run, player, x, y, value, gok, greason, wok, wreason)
					}
				case 2:
					if got.Undo() != want.undo() {
						t.Fatalf("seq %d run %d Undo 结果不符", seq, run)
					}
				}
				if got.Value(x, y) != want.value(x, y) {
					t.Fatalf("seq %d run %d 地块 (%d,%d) 值不符：%d vs %d",
						seq, run, x, y, got.Value(x, y), want.value(x, y))
				}
				if len(got.stack) != len(want.stack) {
					t.Fatalf("seq %d run %d 撤销栈深度不符：%d vs %d",
						seq, run, len(got.stack), len(want.stack))
				}
			}
		}
	}
}
