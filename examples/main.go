package main

import (
	"fmt"
	args "github.com/threeq/cli_args"
	"os"
)

type CompilerOptions struct {
	Module    string `json:"module" usage:"模块"`
	Target    string `json:"target" usage:"目标"`
	SourceMap bool   `json:"sourceMap" usage:"map 文件"`
}

type ExampleConfig struct {
	CompilerOptions CompilerOptions `json:"compilerOptions"`
	Exclude         []string        `json:"exclude"`
}

var conf = &ExampleConfig{
	CompilerOptions: CompilerOptions{Target: "xxx.txt"},
}

func main() {
	app := args.New("args examples", args.Version("0.0.1"),
		args.Store(conf),
		args.Usage("args using demo."),
		args.FileConfigEnabled("config", "example.json", false,
			"config file, support YAML/JSON/TOML. example: --config=example.yaml"),
		args.HelpExit(0),
	)

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("\nError: ")
		fmt.Printf("\n  %s", err.Error())
		fmt.Printf("\n\n")
		os.Exit(1)
	}

	fmt.Printf("获取参数 compilerOptions.module = %v\n", conf.CompilerOptions.Module)
	fmt.Printf("获取参数 compilerOptions.target = %v\n", conf.CompilerOptions.Target)
	fmt.Printf("获取参数 compilerOptions.sourceMap = %v\n", conf.CompilerOptions.SourceMap)
	fmt.Printf("获取参数 exclude = %v\n", conf.Exclude)
}
