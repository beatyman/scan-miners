# Scan Miners

## 简介
这是一个用于扫描 Antpool 矿工和矿机状态的 Golang 应用程序。

## 功能
1.  **Antpool Worker 扫描**: 从 Antpool API 获取 Worker 列表并存入数据库。
2.  **矿机状态扫描**: 根据 Worker ID 生成 IP，扫描矿机本地 API 获取详细状态。

## 运行要求
*   MySQL 数据库
*   Golang 1.22+

## 配置
配置硬编码在 `config/config.go` 中（根据需求）。如需修改数据库连接或 Cookie，请编辑该文件。

## 运行
```bash
# 编译
go build -o sacn-miners.exe ./cmd/main.go

# 运行
./sacn-miners.exe
```

## 项目结构
遵循 Clean Architecture:
*   `cmd/`: 入口
*   `internal/domain/`: 实体和接口
*   `internal/usecase/`: 业务逻辑
*   `internal/repository/`: 数据访问实现
*   `pkg/`: 公共库
