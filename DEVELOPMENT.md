# 项目开发文档 - Scan Miners

## 1. 项目概述

本项目旨在实现一个矿机监控系统，主要功能包括：
1.  从 Antpool 获取 Worker 列表并存储到 MySQL 数据库。
2.  根据 Worker 信息生成 IP 地址，扫描矿机状态接口（Stats），并将数据存储到 MySQL 数据库。

项目采用 Golang 编写，遵循 Clean Architecture（整洁架构）原则。

## 2. 技术栈

*   **语言**: Golang
*   **架构**: Clean Architecture
*   **ORM**: GORM
*   **数据库**: MySQL
*   **日志**: 清晰的结构化日志 (使用 zap 或标准库封装)

## 3. 需求分析

### 3.1 Antpool Worker 数据采集

*   **数据源**: Antpool API
*   **请求方式**: HTTP GET (模拟 curl 请求)
*   **处理逻辑**:
    *   解析 JSON 响应中的 `data.items`。
    *   提取关键字段：`workerId`, `hsLast10Min`, `hsLast1Hour`, `hsLast1D` 等。
    *   **IP 生成规则**:
        *   `workerId` 格式示例: `30x182`
        *   生成 IP: `172.16.30.182` (将 `x` 替换为 `.`, 前缀 `172.16.`)
    *   **数据转换**:
        *   算力字段 (如 `306.23 TH/s`) 需拆分为数值 (`306.23`) 和单位 (`TH/s`)。
*   **存储**: 存入 `workers` 表，`workerId` 为唯一索引。

### 3.2 矿机 Stats 数据采集

*   **数据源**: 矿机本地 API
    *   Endpoint 1: `http://<IP>/cgi-bin/get_stats.cgi`
    *   Endpoint 2: `http://<IP>/cgi-bin/stats.cgi`
    *   策略: 优先尝试 Endpoint 1，失败则尝试 Endpoint 2。
*   **处理逻辑**:
    *   解析返回的 JSON 数据。
    *   **层级解析**:
        *   第一层: `STATUS`, `INFO`, `STATS` (数组)。
        *   第二层: `STATS` 内部的 `chain` (数组)。
    *   **存储结构**:
        *   主表 `miner_stats`: 存储 `INFO` 信息及 `STATS` 数组中非数组字段（打平存储）。
        *   子表 `miner_chains`: 存储 `STATS` -> `chain` 数组中的信息。
*   **关联**: 通过 IP 或 Worker ID 关联 `workers` 表。

## 4. 数据库设计

### 4.1 Workers 表 (`workers`)

| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| id | BIGINT | 主键 |
| worker_id | VARCHAR(64) | 唯一索引, 原始 Worker ID |
| ip | VARCHAR(64) | 生成的 IP 地址 |
| user_worker_id | VARCHAR(128) | 用户 Worker ID |
| worker_status | INT | 状态 |
| hs_last_10min | DOUBLE | 10分钟算力数值 |
| hs_last_10min_unit | VARCHAR(16) | 10分钟算力单位 |
| hs_last_1h | DOUBLE | 1小时算力数值 |
| hs_last_1h_unit | VARCHAR(16) | 1小时算力单位 |
| hs_last_1d | DOUBLE | 24小时算力数值 |
| hs_last_1d_unit | VARCHAR(16) | 24小时算力单位 |
| reject_ratio | VARCHAR(16) | 拒绝率 |
| online_time_last_24h | DOUBLE | 24小时在线时间 |
| created_at | DATETIME | 创建时间 |
| updated_at | DATETIME | 更新时间 |

### 4.2 Miner Stats 表 (`miner_stats`)

| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| id | BIGINT | 主键 |
| worker_id | VARCHAR(64) | 关联 Workers 表 (外键或逻辑关联) |
| ip | VARCHAR(64) | 矿机 IP |
| miner_type | VARCHAR(64) | INFO.type |
| miner_version | VARCHAR(64) | INFO.miner_version |
| compile_time | VARCHAR(64) | INFO.CompileTime |
| elapsed | BIGINT | STATS.elapsed |
| rate_5s | DOUBLE | STATS.rate_5s |
| rate_30m | DOUBLE | STATS.rate_30m |
| rate_avg | DOUBLE | STATS.rate_avg |
| rate_ideal | DOUBLE | STATS.rate_ideal |
| rate_unit | VARCHAR(16) | STATS.rate_unit |
| fan_num | INT | STATS.fan_num |
| hwp_total | DOUBLE | STATS.hwp_total |
| created_at | DATETIME | 创建时间 |

### 4.3 Miner Chains 表 (`miner_chains`)

| 字段名 | 类型 | 说明 |
| :--- | :--- | :--- |
| id | BIGINT | 主键 |
| miner_stat_id | BIGINT | 关联 Miner Stats 表 |
| chain_index | INT | 链索引 (index) |
| freq_avg | INT | 平均频率 |
| rate_ideal | DOUBLE | 理想算力 |
| rate_real | DOUBLE | 实际算力 |
| asic_num | INT | ASIC 数量 |
| hw | INT | 硬件错误数 |
| hwp | DOUBLE | 硬件错误百分比 |
| temp_chip_avg | DOUBLE | 芯片平均温度 (计算值或取第一个) |
| temp_pcb_avg | DOUBLE | PCB 平均温度 |
| created_at | DATETIME | 创建时间 |

*(注：temp_pic, temp_pcb, temp_chip 为数组，为简化存储，可视需求存储其字符串形式或平均值，本设计暂存关键指标或字符串)*

## 5. 项目结构 (Clean Architecture)

```
scan-miners/
├── cmd/
│   └── main.go           # 程序入口
├── config/
│   └── config.go         # 配置加载
├── internal/
│   ├── domain/           # 领域实体 (Entities) 和 接口定义 (Interfaces)
│   │   ├── model/        # 数据模型
│   │   └── repository/   # 仓库接口
│   ├── usecase/          # 业务逻辑 (Use Cases)
│   ├── repository/       # 数据访问层实现 (Repository Implementation)
│   │   └── mysql/        # MySQL 实现
│   └── delivery/         # 外部接口层
│       └── cron/         # 定时任务或一次性执行入口
├── pkg/
│   ├── database/         # 数据库连接封装
│   ├── logger/           # 日志封装
│   └── utils/            # 工具函数
├── go.mod
└── README.md
```
