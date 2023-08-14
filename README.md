# aurora

```shell
    +--------------------------------------------------+
    |          ___                                     |
    |         /   | __  ___________  _________ _       |
    |        / /| |/ / / / ___/ __ \/ ___/ __ `/       |
    |       / ___ / /_/ / /  / /_/ / /  / /_/ /        |
    |      /_/  |_\__,_/_/   \____/_/   \__,_/         |
    |                                                  |
    |                                                  |
    |                >_ Aurora v1.0.0                  |
    |                   MIT License                    |
    |          Copyright (c) 2023 stubborn-gaga        |
    +--------------------------------------------------+
```

> A Golang command-line tool based on the "spf13/cobra" package, designed to manage your "prepare2go" project.
---------

## aurora create <project-name>

> 基于 **[prepare2go](https://github.com/stubborn-gaga-0805/prepare2go)** 创建一个新项目。该命令提供基于 **TUI** 的交互式界面来创建一个新的项目或Demo示例。

```shell
# example:
$ aurora create demo-project
$ aurora create demo-project -p "~/golang/src/demo" --with.demo
```

- 可用选项：
    - **-h, --help**  查看帮助信息
    - **-p, --path**  指定项目路径 ( 选择已存在的路径可能会被覆盖 )
    - **--with.demo** 是否创建Demo项目 ( 非本地开发环境禁用此选项 )

## aurora init

> 初始化项目。对项目的包依赖、必要的命令行工具进行初始化和安装

```shell
# example:
$ aurora init
```

- 可用选项：
    - **-h, --help**  查看帮助信息

## aurora build

> 编译当前项目。默认输出二进制文件为: ```./bin/server```

```shell
# example:
$ aurora build
```

- 可用选项：
    - **-h, --help**  查看帮助信息
    - **-o, --output**   自定义二进制文件输出路径

## aurora gen-model

> 使用基于 **TUI** 的交互式界面来生成```model```文件

```shell
# example:
$ aurora gen-model  # 进入交互界面
$ aurora gen-model -c "db" -t "table_a,table_b"
```

- 可用选项：
    - **-h, --help**  查看帮助信息
    - **-c, --conn**  配置文件中的连接配置，默认: "db"
    - **-o, --output**  执行生成文件的路径,默认: "./internal/repo/orm"
    - **-p, --pkg** 生成model文件的包名,默认: "orm", 需要和生成路径的文件夹对应
    - **-t, --table** 指定生成的表名 (多张表用","隔开)

## aurora run

> 启动项目。该命令每次执行都会自动编译创建二进制文件:  ```./bin/server```

```shell
# example:
$ aurora run  # 编译并启动项目
$ aurora run -e dev --with.corn # 用dev的环境配置编译并启动项目，并且启动crontab任务
$ aurora run -e dev --without.mq # 用dev的配置编译并启动项目，不启动mq
```

- 可用选项：
    - **-h, --help**  查看帮助信息
    - **-c, --config**  设置配置文件的路径
    - **-e, --env** 设置服务的运行环境 (默认: "local")
    - **-n, --name**  设置服务名称 (默认: "prepare-to-go")
    - **-v, --version** 设置应用的版本 (默认: "v1.0")
    - **--with.cron** 是否启动定时任务
    - **--with.ws** 是否启动websocket服务
    - **--without.mq** 不启动MQ
    - **--without.server** 不启动http服务

## aurora job <job-name>

> 执行用户自定义脚本任务。

```shell
# example:
$ aurora job test-job -p "1,2" # 执行任务 test-job，将1、2两个参数传给任务函数
```

- 可用选项:
    - **-h, --help**  查看帮助信息
    - **-l, --list**  查看可执行的用户任务
    - **-p, --params**  运行命令的参数, 多个参数用","隔开

## aurora cron

> 定时任务相关

```shell
# example:
$ aurora cron -l  # 查看运行中的crontab任务
```

- 可用选项：
    - **-h, --help**  查看帮助信息
    - **-l, --list**  查看运行中的crontab任务
