package terraedit

import "testing"

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
