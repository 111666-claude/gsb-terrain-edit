// Command terrain-edit 跑地块编辑样例。
package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/terraedit"
)

// Run 执行一次命令行调用，返回退出码。
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("terrain-edit", flag.ContinueOnError)
	flags.SetOutput(stderr)
	sample := flags.String("sample", "permission", "permission / duplicate / undo")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	if *sample == "permission" {
		editor := terraedit.NewEditor(map[string]string{"p": "guest"})
		ok, reason := editor.Apply("p", 1, 1, 5)
		fmt.Fprintf(stdout, "ok=%v reason=%s\n", ok, reason)
		return 0
	}
	if *sample == "duplicate" {
		editor := terraedit.NewEditor(map[string]string{"b": "builder"})
		editor.Apply("b", 1, 1, 5)
		ok, reason := editor.Apply("b", 1, 1, 5)
		fmt.Fprintf(stdout, "second=%v reason=%s\n", ok, reason)
		return 0
	}
	if *sample == "undo" {
		editor := terraedit.NewEditor(map[string]string{"b": "builder"})
		editor.Apply("b", 1, 1, 5)
		ok := editor.Undo()
		fmt.Fprintf(stdout, "ok=%v value=%d\n", ok, editor.Value(1, 1))
		return 0
	}
	fmt.Fprintln(stderr, "需要 --sample permission|duplicate|undo")
	return 2
}

func main() {
	os.Exit(Run(os.Args[1:], os.Stdout, os.Stderr))
}
