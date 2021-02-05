package args

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
)

/*
1. 支持 yaml/json/toml
2. 优先级：arg > env > file
3. 自动解析 obj 到 arg 参数格式
*/

var ErrHelp = errors.New("flag: help requested")
var ErrCmdParse = errors.New("参数解析错误")
var ErrFileNotFound = errors.New("文件未找到")
var ErrFileRead = errors.New("文件读取失败")
var ErrFileParse = errors.New("文件解析错误")
var ErrFileType = errors.New("文件类型不支持。仅支持：【json/yml/yaml/toml/ini】")
var ErrArgType = errors.New("不支持的参数类型")

// AppArgs 参数解析应用类型
type AppArgs struct {
	Name           string
	Version        string
	Usage          string
	CfgData        interface{}
	CfgFilePath    string
	CfgFileCmdArg  string
	CfgFileUsage   string
	CfgFileRequire bool
	EnvPrefix      string
	HelpHandler    func() error
	output         io.Writer
}

// Run 运行参数解析
func (a *AppArgs) Run(arguments []string) error {
	flags := Bean2Args(a.CfgData)
	set := a.flagSet(a.Name, flags)
	if a.CfgFileCmdArg != "" {
		set.String(a.CfgFileCmdArg, a.CfgFilePath, a.CfgFileUsage)
	}

	if err := set.Parse(arguments[1:]); err != nil {
		if err == flag.ErrHelp {
			if a.HelpHandler != nil {
				return a.HelpHandler()
			}
		} else {
			fmt.Printf("参数解析错误: %v\n", err)
			return ErrCmdParse
		}
	}

	// 处理 配置文件 参数
	err := a.parseFileArg(set)
	if err != nil && a.CfgFileRequire {
		return err
	}
	// 处理 环境变量 参数
	a.parseEnvArg(set, flags)
	// 处理 命令行   参数
	a.parseCmdArg(set, flags)

	return nil
}

// parseFileArg 解析配置文件参数
func (a *AppArgs) parseFileArg(set *flag.FlagSet) error {
	if a.CfgFileCmdArg == "" {
		return nil
	}
	cfgFlag := set.Lookup(a.CfgFileCmdArg)
	if cfgFlag.Value.String() != "" {
		a.CfgFilePath = cfgFlag.Value.String()
	}
	a.CfgFilePath = strings.TrimSpace(a.CfgFilePath)
	if a.CfgFilePath == "" {
		return errors.New("flag required but not provided: -" + a.CfgFileCmdArg)
	}

	extName := path.Ext(a.CfgFilePath)[1:]
	var unmarshal func(data []byte, v interface{}) error
	if extName == "json" {
		unmarshal = json.Unmarshal
	} else if extName == "yml" || extName == "yaml" {
		unmarshal = yaml.Unmarshal
	} else if extName == "toml" || extName == "ini" {
		unmarshal = toml.Unmarshal
	} else {
		return ErrFileType
	}

	fmt.Fprintf(set.Output(), "读取配置文件 %s\n", a.CfgFilePath)
	file, err := os.Open(a.CfgFilePath)
	if err != nil {
		_, _ = fmt.Fprintf(set.Output(), "配置文件[%s]打开错误：%v\n", a.CfgFilePath, err)
		return ErrFileNotFound
	}

	content, err := ioutil.ReadAll(file)
	if err != nil {
		_, _ = fmt.Fprintf(set.Output(), "配置文件[%s]读取错误：%v\n", a.CfgFilePath, err)
		return ErrFileRead
	}

	err = unmarshal(content, a.CfgData)
	if err != nil {
		_, _ = fmt.Fprintf(set.Output(), "配置文件[%s]解析错误：%v\n", a.CfgFilePath, err)
		return ErrFileParse
	}
	return nil
}

func getEnvName(name string) {

}

// parseEnvArg 解析环境变量参数
func (a *AppArgs) parseEnvArg(set *flag.FlagSet, flags map[string]*StructArg) {
	for name, f := range flags {
		envName := a.getEnvName(name)

		envValue, found := os.LookupEnv(envName)
		if !found {
			continue
		}
		v, err := typeValue(f, envValue)
		if err != nil {
			_, _ = fmt.Fprintf(set.Output(), "参数【%v=%v】解析错误：%v\n", envName, envValue, err)
			continue
		}
		f.Set(v)
	}
}

// 解析命令行参数
func (a *AppArgs) parseCmdArg(set *flag.FlagSet, flags map[string]*StructArg) {
	set.Visit(func(f *flag.Flag) {
		ff, found := flags[f.Name]
		if !found {
			return
		}
		argValue := f.Value.String()
		if argValue == "" {
			argValue = f.DefValue
		}
		if argValue == "" {
			return
		}

		v, err := typeValue(ff, argValue)
		if err != nil {
			_, _ = fmt.Fprintf(set.Output(), "参数【%v=%v】解析错误：%v\n", f.Name, argValue, err)
			return
		}
		ff.Set(v)
	})
}

// flagSet 参数解析 FlagSet
func (a *AppArgs) flagSet(name string, flags map[string]*StructArg) *flag.FlagSet {
	set := flag.NewFlagSet(name, flag.ContinueOnError)
	set.SetOutput(a.output)
	set.Usage = func() {
		if set.Name() == "" {
			_, _ = fmt.Fprintf(set.Output(), "Usage")
		} else {
			_, _ = fmt.Fprintf(set.Output(), "Usage of %s", set.Name())
		}
		if a.Version == "" {
			_, _ = fmt.Fprintf(set.Output(), ":\n")
		} else {
			_, _ = fmt.Fprintf(set.Output(), "@%s:\n", a.Version)
		}
		if a.Usage != "" {
			_, _ = fmt.Fprintf(set.Output(), "  %s\n", a.Usage)
		}
		argsUsagePrefix := "        "
		_, _ = fmt.Fprintf(set.Output(), "\n  -h, -help\n")
		_, _ = fmt.Fprintf(set.Output(), argsUsagePrefix+"show usage\n")
		set.VisitAll(func(f *flag.Flag) {
			s := fmt.Sprintf("  -%s", f.Name) // Two spaces before -; see next two comments.
			tName, usage := flag.UnquoteUsage(f)
			ff, found := flags[f.Name]
			if found {
				tName = typeName(ff)
			}
			if len(tName) > 0 {
				s += " " + tName
			}
			// Env name
			s += fmt.Sprintf(" \t (ENV: %s)", a.getEnvName(f.Name))
			// Boolean flags of one ASCII letter are so common we
			// treat them specially, putting their usage on the same line.
			if len(s) <= 4 { // space, space, '-', 'x'.
				s += argsUsagePrefix + ""
			} else {
				// Four spaces before the tab triggers good alignment
				// for both 4- and 8-space tab stops.
				s += "\n" + argsUsagePrefix
			}
			s += strings.ReplaceAll(usage, "\n", "\n"+argsUsagePrefix)

			if !isZeroValue(f, f.DefValue) {
				if tName == "string" {
					// put quotes on the value
					s += fmt.Sprintf(" (default %q)", f.DefValue)
				} else {
					s += fmt.Sprintf(" (default %v)", f.DefValue)
				}
			}
			fmt.Fprint(set.Output(), s, "\n")
		})
	}
	for _, f := range flags {
		_ = set.String(f.Name, f.Default, f.Usage)
	}
	return set
}

func (a *AppArgs) getEnvName(name string) string {
	envName := strings.ReplaceAll(name, ".", "_")
	if a.EnvPrefix != "" {
		envName = a.EnvPrefix + "_" + envName
	}
	return strings.ToUpper(envName)
}

// New 新建参数解析应用
func New(name string, options ...Option) *AppArgs {

	args := &AppArgs{
		Name: name,
		HelpHandler: func() error {
			return ErrHelp
		},
	}

	for _, opt := range options {
		opt(args)
	}

	return args
}

// Option AppArgs 配置类型
type Option func(args *AppArgs)

// Store 参数存储对象
func Store(config interface{}) Option {
	return func(args *AppArgs) {
		args.CfgData = config
	}
}

func Version(version string) Option {
	return func(args *AppArgs) {
		args.Version = version
	}
}
func Usage(usage string) Option {
	return func(args *AppArgs) {
		args.Usage = usage
	}
}

func HelpExit(code int) Option {
	return func(args *AppArgs) {
		args.HelpHandler = func() (err error) { os.Exit(code); return }
	}
}

func Output(output io.Writer) Option {
	return func(args *AppArgs) {
		args.output = output
	}
}

// 文件配置参数
func FileConfigEnabled(argName, defaultValue string, require bool, usage string) Option {
	return func(args *AppArgs) {
		args.CfgFileCmdArg = argName
		args.CfgFilePath = defaultValue
		args.CfgFileRequire = require
		if usage == "" {
			args.CfgFileUsage = "配置文件"
		} else {
			args.CfgFileUsage = usage
		}
	}
}

// 环境变量配置
func EnvArg(prefix string) Option {
	return func(args *AppArgs) {
		args.EnvPrefix = prefix
	}
}

// isZeroValue determines whether the string represents the zero
// value for a flag.
func isZeroValue(f *flag.Flag, value string) bool {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(f.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Ptr {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	return value == z.Interface().(flag.Value).String()
}

func typeValue(f *StructArg, envValue string) (interface{}, error) {
	switch f.T {
	case reflect.Bool:
		return strconv.ParseBool(envValue)
	case reflect.Int, reflect.Int32:
		return strconv.Atoi(envValue)
	case reflect.Int8:
		var i, err = strconv.ParseInt(envValue, 10, 8)
		return int8(i), err
	case reflect.Int16:
		var i, err = strconv.ParseInt(envValue, 10, 16)
		return int16(i), err
	case reflect.Int64:
		return strconv.ParseInt(envValue, 10, 64)
	case reflect.Uint, reflect.Uint32:
		var i, err = strconv.ParseUint(envValue, 10, 32)
		return uint(i), err
	case reflect.Uint8:
		var i, err = strconv.ParseUint(envValue, 10, 8)
		return uint8(i), err
	case reflect.Uint16:
		var i, err = strconv.ParseUint(envValue, 10, 16)
		return uint16(i), err
	case reflect.Uint64:
		return strconv.ParseUint(envValue, 10, 64)
	case reflect.Float32:
		var f, err = strconv.ParseFloat(envValue, 32)
		return float32(f), err
	case reflect.Float64:
		return strconv.ParseFloat(envValue, 64)
	case reflect.String:
		return envValue, nil
	default:
		return nil, ErrArgType
	}
}

func typeName(t *StructArg) string {
	return t.TName
}

// StructArg 结构参数
type StructArg struct {
	T       reflect.Kind
	Name    string
	Default string
	Usage   string
	Require bool
	Set     func(value interface{})
	TName   string
}

// Bean2Args 对象到 AppArgs 转换
func Bean2Args(data interface{}) map[string]*StructArg {
	args := map[string]*StructArg{}

	t, _ := realTV(reflect.TypeOf(data), reflect.ValueOf(data))
	if t.Kind() != reflect.Struct {
		panic("不支持普通类型参数解析，请使用结构类型接收参数！！！")
	}
	bean2XPath(args, reflect.TypeOf(data), reflect.ValueOf(data), "", false, "")

	return args
}

// bean2XPath 得到对象的 xpath
func bean2XPath(args map[string]*StructArg, t reflect.Type, v reflect.Value, path string, require bool, usage string) {

	t, v = realTV(t, v)

	switch t.Kind() {
	case reflect.Struct:
		numField := t.NumField()
		for i := 0; i < numField; i++ {
			field := t.Field(i)

			if argName, found := tagArgName(field); found {
				if path != "" {
					argName = path + "." + argName
				}
				usage := field.Tag.Get("usage")
				_, require := field.Tag.Lookup("require")
				bean2XPath(args, field.Type, v.Field(i), argName, require, usage)
			}
		}
	case reflect.Complex64,
		reflect.Complex128,
		reflect.Array,
		reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Uintptr,
		reflect.Ptr,
		reflect.Slice,
		reflect.UnsafePointer:
		// 不支持数据类型
		return
	default:
		args[path] = &StructArg{
			T:       t.Kind(),
			TName:   t.Name(),
			Name:    path,
			Usage:   usage,
			Default: fmt.Sprintf("%v", v),
			Require: require,
			Set: func(value interface{}) {
				v.Set(reflect.ValueOf(value))
			},
		}
	}
}

// tagArgName 获取 tag 配置的参数信息
func tagArgName(field reflect.StructField) (string, bool) {
	if name, found := field.Tag.Lookup("yaml"); found {
		return name, found
	} else if name, found := field.Tag.Lookup("json"); found {
		return name, found
	} else if name, found := field.Tag.Lookup("toml"); found {
		return name, found
	}
	return "", false
}

// realType 得到指针类型的真实类型
func realTV(t reflect.Type, v reflect.Value) (reflect.Type, reflect.Value) {
	for ; t.Kind() == reflect.Ptr; {
		t = t.Elem()
		v = v.Elem()
	}

	return t, v
}
