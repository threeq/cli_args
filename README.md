# cli_args
go 命令行参数解析库。特点支持对象自动解析成命令行参数

# Features
- [x] 对象自动解析成参数
- [x] 命令行参数
- [x] 环境变量参数
- [x] 文件参数
- [x]（优先级：命令行 > 环境变量 > 文件）

# Use

安装库

```bash
go get "github.com/threeq/cli_args"
```

程序里使用

```go
package main

import args "github.com/threeq/cli_args"
import "fmt"
import "os"

var conf = new(struct{
    // more your configure define
    Demo string `json:"demo"`
})

func main()  {
    app := args.New("Application Name", 
        args.Version("0.0.1"),
 		args.Store(conf),  // ** 存储对象必须是 struct 结构体 **
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
}
```

详细 demo 请往 [examples](examples) 查看 