package console

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"modstart-cli/lib/api"
)

// Prompt 读取一行输入
func Prompt(label string) string {
	fmt.Print(label)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// PrintJSON 输出 JSON（中文不转义、HTML 字符不转义）
func PrintJSON(v interface{}) {
	if raw, ok := v.(json.RawMessage); ok {
		var x interface{}
		if err := json.Unmarshal(raw, &x); err == nil {
			v = x
		}
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		fmt.Println(v)
	}
}

// PrintApis 文字输出接口列表
func PrintApis(apis []api.ApiDef) {
	if len(apis) == 0 {
		fmt.Println("当前站点没有注册任何 CLI 接口")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\t说明\t参数")
	for _, a := range apis {
		var params []string
		for _, p := range a.Params {
			name := p.Name
			if p.Required {
				name += "*"
			}
			params = append(params, name)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", a.Key, a.Title, strings.Join(params, ","))
	}
	w.Flush()
	fmt.Println("\n说明：参数名后带 * 为必填；调用示例：modstart-cli call <key> '{\"page\":1}'")
}
