# 数据库设计文档

简历优化系统的完整数据库设计，包括 MySQL（使用 Ent ORM）和 MongoDB 存储方案。

---

## 目录

1. [MySQL 表设计](#mysql-表设计)
2. [Ent Schema 定义](#ent-schema-定义)
3. [MongoDB 存储设计](#mongodb-存储设计)
4. [数据关系说明](#数据关系说明)
5. [使用示例](#使用示例)

---

## MySQL 表设计

### 1. 简历表 (resume) - `cv_resume`

**说明**：存储简历基本信息

**字段**：
- `id` (int64, 雪花ID) - 主键
- `user_id` (int64) - 用户ID
- `tenant_id` (int64) - 租户ID
- `file_path` (string) - 文件路径
- `file_name` (string) - 文件名
- `status` (int32) - 状态: 1=pending, 2=processing, 3=completed
- `created_at` (time) - 创建时间
- `updated_at` (time) - 更新时间
- `deleted_at` (time) - 删除时间（软删除）

**关系**：
- 一对多关联到 `ResumeScore`

**索引**：
- `(user_id, tenant_id)`
- `(status)`

---

### 2. 模块表 (module) - `cv_module`

**说明**：定义简历评分的模块类型（如：基本信息、工作经历、教育背景、项目经验等）

**字段**：
- `id` (int64, 雪花ID) - 主键
- `title` (string) - 显示标题
- `description` (string) - 模块描述
- `created_at` (time) - 创建时间
- `updated_at` (time) - 更新时间
- `deleted_at` (time) - 删除时间（软删除）

**关系**：
- 一对多关联到 `Dimension`

---

### 3. 维度表 (dimension) - `cv_dimension`

**说明**：定义每个模块下的评分维度（如：完整性、准确性、相关性等）

**字段**：
- `id` (int64, 雪花ID) - 主键
- `module_id` (int64) - 所属模块ID
- `title` (string) - 显示标题
- `judgment` (JSON) - 维度得分详情数组，每项包含 `{detail, score, weight}`
- `created_at` (time) - 创建时间
- `updated_at` (time) - 更新时间
- `deleted_at` (time) - 删除时间（软删除）

**关系**：
- 多对一关联到 `Module`

**索引**：
- `(module_id)`

**judgment 字段示例**：
```json
[
  {
    "detail": "简历包含完整的联系方式",
    "score": 10,
    "weight": 0.3
  },
  {
    "detail": "工作经历描述详细",
    "score": 8,
    "weight": 0.7
  }
]
```

---

### 4. 简历得分表 (resume_score) - `cv_resume_score`

**说明**：存储每份简历在各个模块或维度的得分（多态关联设计）

**字段**：
- `id` (int64, 雪花ID) - 主键
- `resume_id` (int64) - 关联简历ID
- `target_id` (int64) - 关联的模块ID或维度ID
- `target_type` (int32) - 类型: 0=module, 1=dimension
- `score` (float64) - 得分
- `weight` (float64) - 权重
- `created_at` (time) - 创建时间
- `updated_at` (time) - 更新时间
- `deleted_at` (time) - 删除时间（软删除）

**关系**：
- 多对一关联到 `Resume`

**索引**：
- `(resume_id, target_type, target_id)`
- `(target_type, target_id)`

**多态关联说明**：
- 当 `target_type = 0` 时，`target_id` 指向 `Module.id`
- 当 `target_type = 1` 时，`target_id` 指向 `Dimension.id`

---

## Ent Schema 定义

### Schema 文件位置

```
internal/infra/ent/schema/
├── resume.go          # 简历表 schema
├── module.go          # 模块表 schema
├── dimension.go       # 维度表 schema
└── resume_score.go    # 得分表 schema
```

### 生成 Ent 代码

```bash
make generate
```

或者：

```bash
go generate ./internal/infra/ent/...
```

---

## MongoDB 存储设计

### Collection: `resume_content`

**用途**：存储简历原始内容和解析后的结构化数据

**字段**：
- `_id` (ObjectId) - MongoDB 主键
- `mysql_id` (int64) - 关联 MySQL 简历表的 ID
- `create_time` (timestamp) - 创建时间
- `update_time` (timestamp) - 更新时间
- `raw_md` (string) - 原始 Markdown 内容
- `modules` (array) - 模块化数据数组

### ModuleData 结构

每个模块包含：
- `module_id` (int64) - 对应 MySQL `module.id`
- `title` (string) - 模块标题
- `data` (array) - 模块具体数据（灵活结构）

### MongoDB Go 模型

位置：`internal/infra/mongo/model/resumecontentmodel.go`

```go
type ResumeContent struct {
    ID         primitive.ObjectID `bson:"_id,omitempty"`
    MySQLID    int64              `bson:"mysql_id"`
    CreateTime time.Time          `bson:"create_time"`
    UpdateTime time.Time          `bson:"update_time"`
    RawMD      string             `bson:"raw_md"`
    Modules    []ModuleData       `bson:"modules"`
}

type ModuleData struct {
    ModuleID int64                    `bson:"module_id"`
    Title    string                   `bson:"title"`
    Data     []map[string]interface{} `bson:"data"`
}
```

---

## 数据关系说明

### MySQL 表关系

```
Resume (简历)
  └── ResumeScore (简历得分) [1:N]
        ├── Module (模块) [多态关联]
        └── Dimension (维度) [多态关联]

Module (模块)
  └── Dimension (维度) [1:N]
```

### MySQL 与 MongoDB 关联

```
MySQL (cv_resume)          MongoDB (resume_content)
├── id (int64)      ←─────  mysql_id (int64)
├── user_id
├── file_path
└── status

MySQL (cv_module)          MongoDB (ModuleData)
├── id (int64)      ←─────  module_id (int64)
└── title           ←─────  title (string)
```

---

## 项目文件结构

```
cv2/
├── etc/
│   └── cv2.yaml                          # 配置文件
├── internal/
│   ├── config/
│   │   └── config.go                     # Config 结构体
│   ├── infra/                            # 基础设施层
│   │   ├── ent/                          # Ent ORM
│   │   │   ├── schema/
│   │   │   └── [生成的代码...]
│   │   ├── mongo/                        # MongoDB 客户端
│   │   │   ├── client.go
│   │   │   └── model/
│   │   └── minio/                        # MinIO 客户端
│   │       └── client.go
│   └── svc/
│       ├── db.go                         # MySQL 初始化
│       ├── mongo.go                      # MongoDB 初始化
│       ├── minio.go                      # MinIO 初始化
│       └── serviceContext.go             # 服务上下文
└── docs/
    └── db.md                             # 本文档
```

---

## 常用命令

```bash
# 生成 Ent 代码
make generate

# 编译项目
make build

# 运行项目
make run

# 查看帮助
make help
```