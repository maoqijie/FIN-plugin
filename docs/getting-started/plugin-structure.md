## 目录约定

主程序运行时会在工作目录创建单一插件目录，所有插件按名称归档：

```
Plugin/grpc/
  example/
    example.so
    plugin.yaml
    assets/
```

在本仓库中，`templates/` 提供示例骨架，`docs/` 维护文档，`sdk/` 存放对接主程序的公共接口定义。

## 插件结构

每个插件需要放在 `Plugin/grpc/<插件名>/` 目录下，例如：

```
Plugin/grpc/example/
  main.so             # 默认入口文件
  plugin.yaml         # 可选，补充元数据
  assets/
  README.md
```

### `plugin.yaml` Manifest

Manifest 用于描述插件的基本信息及运行入口，建议使用 YAML：

```yaml
name: example
displayName: 示例插件
version: 0.1.0
entry: ./bin/example.so   # 可选，缺省为 main.so
sdkVersion: 0.1.0
authors:
  - 猫七街
description: |
  这是一个演示插件，展示如何监听游戏事件并广播到 QQ。
dependencies:
  - name: core
    version: ">=0.1.0"
permissions:
  - minecraft.chat.read
  - minecraft.chat.write
  - qq.group.send
config:
  enable: true
  targetGroup: 123456789
```

字段说明：

- `name`：插件唯一标识。
- `entry`：入口脚本或可执行文件。缺省时主程序会使用 `main.so`。
- `sdkVersion`：声明依赖的 SDK 版本，便于主程序做兼容检查。
- `dependencies`：插件间依赖（可选）。
- `permissions`：声明本插件需要访问的能力，便于后续统一治理。
- `config`：插件默认配置，主程序首次加载时可据此生成用户可编辑的配置文件。

当插件目录中存在 `.go` 源码时，主程序会在加载或热重载阶段自动执行 `go build -buildmode=plugin -o main.so .` 生成共享库（目前仅支持 Linux 与 macOS）；因此只需提交源码即可，标题所述的 `main.so` 会由运行实例按需编译。

## 生命周期约定