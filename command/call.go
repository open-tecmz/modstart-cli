package command

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"modstart-cli/lib/api"
	"modstart-cli/lib/config"
	"modstart-cli/lib/console"
)

var flagData string

var callCmd = &cobra.Command{
	Use:   "call <key> [json]",
	Short: "调用站点已注册的后台接口",
	Long: `调用站点已注册的后台接口，参数统一使用 JSON 对象。

示例：
  modstart-cli call news.index '{"page":1,"pageSize":20}'
  modstart-cli call news.add '{"title":"标题","content":"内容"}'
  modstart-cli call news.category.index
  modstart-cli call news.delete '{"ids":[4,5]}'

参数也可以放在 --data 中（与第二个位置参数等价）：
  modstart-cli call news.edit --data '{"id":1,"title":"新标题"}'

默认输出统一 JSON（中文不转义），参数会按接口定义在本地校验。`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		raw := strings.TrimSpace(flagData)
		if len(args) > 1 {
			if raw != "" {
				return fmt.Errorf("参数 JSON 只能提供一次：使用位置参数或 --data 其中之一")
			}
			raw = args[1]
		}
		rawParams, err := parseParamsJSON(raw)
		if err != nil {
			return err
		}

		site, err := activeSite()
		if err != nil {
			return err
		}
		client := api.New(site)
		list, err := client.ApiList()
		if err != nil {
			return err
		}

		var def *api.ApiDef
		for i := range list.Apis {
			if list.Apis[i].Key == key {
				def = &list.Apis[i]
				break
			}
		}
		if def == nil {
			return fmt.Errorf("未找到接口: %s（可执行 modstart-cli api list 查看）", key)
		}

		params := normalizeParams(def, rawParams)
		// 本地校验：必填、类型、未声明参数提示
		if err := validateParams(def, params); err != nil {
			return err
		}

		resp, err := client.Call(def.URL, def.Method, params)
		if err != nil {
			return err
		}
		if err := config.Save(cfg); err != nil {
			return err
		}

		console.PrintJSON(resp.View())
		return nil
	},
}

func init() {
	callCmd.Flags().StringVar(&flagData, "data", "", "以 JSON 提供参数（等价于第二个位置参数）")
}

// parseParamsJSON 解析 JSON 对象参数
func parseParamsJSON(raw string) (map[string]interface{}, error) {
	params := map[string]interface{}{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return params, nil
	}
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&params); err != nil {
		return nil, fmt.Errorf("参数必须是 JSON 对象，如 '{\"page\":1}': %w", err)
	}
	var extra interface{}
	if err := dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("参数 JSON 后存在多余内容")
	}
	return params, nil
}

// normalizeParams 按接口定义把 JSON 值转换为表单可提交的字符串
//
// - json / array / object 类型：保留 JSON 字符串
// - 数组且参数为普通类型（如 ids）：用逗号连接，兼容后台 parseIds()
// - 数字、布尔保持字面量
func normalizeParams(def *api.ApiDef, raw map[string]interface{}) map[string]string {
	types := make(map[string]string, len(def.Params))
	for _, p := range def.Params {
		types[p.Name] = strings.ToLower(p.Type)
	}
	params := make(map[string]string, len(raw))
	for k, v := range raw {
		switch t := v.(type) {
		case []interface{}:
			switch types[k] {
			case "json", "array", "object":
				params[k] = jsonValueToString(t)
			default:
				if scalars, ok := joinScalars(t); ok {
					params[k] = scalars
				} else {
					params[k] = jsonValueToString(t)
				}
			}
		case map[string]interface{}:
			params[k] = jsonValueToString(t)
		default:
			params[k] = jsonValueToString(t)
		}
	}
	return params
}

// joinScalars 将标量数组转换为逗号分隔字符串
func joinScalars(items []interface{}) (string, bool) {
	values := make([]string, 0, len(items))
	for _, item := range items {
		switch item.(type) {
		case string, json.Number, bool, nil:
			values = append(values, jsonValueToString(item))
		default:
			return "", false
		}
	}
	return strings.Join(values, ","), true
}

// jsonValueToString 将 JSON 值转换为表单字段值
func jsonValueToString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case json.Number:
		return t.String()
	case bool:
		if t {
			return "true"
		}
		return "false"
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return fmt.Sprintf("%v", t)
		}
		return string(b)
	}
}

// validateParams 根据接口定义在本地校验参数，避免请求发出后才报错
func validateParams(def *api.ApiDef, params map[string]string) error {
	defs := make(map[string]api.ParamDef, len(def.Params))
	for _, p := range def.Params {
		defs[p.Name] = p
	}

	for name, value := range params {
		p, ok := defs[name]
		if !ok {
			fmt.Fprintf(os.Stderr, "提示：接口 %s 未声明参数 %s，仍会发送\n", def.Key, name)
			continue
		}
		switch strings.ToLower(p.Type) {
		case "int", "integer", "bigint", "smallint", "tinyint":
			if _, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64); err != nil {
				return fmt.Errorf("参数 %s 需要整数（%s），当前为 %q", name, p.Desc, value)
			}
		case "float", "double", "decimal", "number":
			if _, err := strconv.ParseFloat(strings.TrimSpace(value), 64); err != nil {
				return fmt.Errorf("参数 %s 需要数字（%s），当前为 %q", name, p.Desc, value)
			}
		case "bool", "boolean":
			v := strings.ToLower(strings.TrimSpace(value))
			if v != "0" && v != "1" && v != "true" && v != "false" {
				return fmt.Errorf("参数 %s 需要布尔值（0/1/true/false），当前为 %q", name, value)
			}
		case "json", "array", "object":
			var x interface{}
			if err := json.Unmarshal([]byte(value), &x); err != nil {
				return fmt.Errorf("参数 %s 需要合法 JSON（%s）", name, p.Desc)
			}
		}
	}

	for _, p := range def.Params {
		if !p.Required {
			continue
		}
		v, ok := params[p.Name]
		if !ok || strings.TrimSpace(v) == "" {
			desc := p.Desc
			if desc == "" {
				desc = p.Name
			}
			return fmt.Errorf("缺少必填参数: %s (%s)", p.Name, desc)
		}
	}
	return nil
}
