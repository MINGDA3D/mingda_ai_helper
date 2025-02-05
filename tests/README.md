# 测试目录结构

本目录包含了项目的集成测试和示例代码。

## 目录说明

- `integration/`: 集成测试
  - `ai_test.go`: AI服务的集成测试
  - `predict_test.go`: 预测功能的集成测试

- `examples/`: 示例代码
  - `moonraker_example.go`: Moonraker客户端使用示例

## 运行测试

### 运行所有集成测试

```bash
go test ./tests/integration/...
```

### 运行特定测试

```bash
# 运行AI测试
go test ./tests/integration -run TestAIService

# 运行预测测试
go test ./tests/integration -run TestPredict
```

### 运行示例

```bash
# 运行Moonraker示例
go run ./tests/examples/moonraker_example.go
```

## 注意事项

1. 集成测试需要确保相关服务（数据库、AI服务等）已经启动
2. 测试配置文件路径使用相对路径 `../../config/config.yaml`
3. 示例代码主要用于演示API的使用方法，不应用于生产环境 