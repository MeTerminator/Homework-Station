# 作业驿站

基于 **TypeScript + React + shadcn/ui** 和 **Gin + GORM** 的作业提交与收集系统，支持 SQLite、MySQL、PostgreSQL。

## 首次启动行为

首次启动会创建数据库表、必要的站点默认配置和一个管理员：

- 用户名：`admin`
- 初始密码：`123456`

初始化信息会同时写入控制台和 `DATA_DIR/app.log`。登录后请修改初始密码。

**初始化不创建任何分类、作业、提交记录、附件或其他示例账号。** 没有预置数据集、演示数据库或自动填充脚本。已有数据库再次启动时保留已有内容、账号密码与配置，不重置为初始状态。

**实名认证默认关闭，认证 API 地址为空。** 管理员配置自己的接口后才能启用，不会自动请求任何预置的第三方认证服务。

## 快速开始

需要 Go 1.25+、Node.js 22.12+、npm 和 C 编译器。SQLite 驱动与 WebP 编码使用 CGO；macOS 需安装 Xcode Command Line Tools，Debian/Ubuntu 可安装 `build-essential`。

在项目根目录执行：

```bash
make install
make build
./backend/bin/homework-station
```

- 用户前台：[http://localhost:8080](http://localhost:8080)
- 管理后台：[http://localhost:8080/admin](http://localhost:8080/admin)

`make build` 先构建前端，再通过 Go `embed` 将全部页面、JavaScript、CSS 等资源编译进 `backend/bin/homework-station`。部署时只需这个可执行文件，无需复制前端目录；运行数据默认保存在启动目录下的 `./data/`。退出时按 `Ctrl+C`。

开发时使用两个终端，分别在项目根目录运行：

```bash
make dev-backend
```

```bash
make dev-web
```

开发页面地址为 [http://localhost:5173](http://localhost:5173)，Vite 将 `/api` 请求代理到 8080。`make dev-backend` 使用 `dev` 构建标签，不要求预先构建前端；默认的 Go 构建会嵌入前端，需要先执行 `cd web && npm run build`。

## 下载发布包与手动发布

在 GitHub 仓库的 **Actions → Release → Run workflow** 中选择要发布的分支，输入版本号（例如 `1.2.3`、`v1.2.3` 或 `1.2.3-rc.1`），点击 **Run workflow**。工作流必须先合入默认分支，才会显示手动运行入口。

工作流先执行完整 CI 检查，再构建前端并嵌入启用 CGO 的各平台后端；所有平台的测试与构建成功后，自动创建指向本次构建提交的 `v版本号` 标签和 GitHub Release，生成更新说明并上传各平台的单文件程序及 `SHA256SUMS`。已有标签会被拒绝，预发布版本自动标记为 Pre-release；无需配置额外 Token，工作流使用内置 `GITHUB_TOKEN`，仅发布任务具有 `contents: write` 权限。

| 平台    | 架构          | 文件名（以 v1.2.3 为例）                                                      | 运行要求                                            |
| ------- | ------------- | ----------------------------------------------------------------------------- | --------------------------------------------------- |
| Linux   | amd64 / arm64 | `homework-station-v1.2.3-linux-amd64` / `homework-station-v1.2.3-linux-arm64` | glibc 2.35+，如 Ubuntu 22.04+；不适用于 Alpine/musl |
| Windows | amd64         | `homework-station-v1.2.3-windows-amd64.exe`                                   | 64 位 Windows                                       |
| macOS   | arm64         | `homework-station-v1.2.3-darwin-arm64`                                        | Apple Silicon，macOS 15+；二进制未经签名或公证      |

下载对应平台的单个可执行文件即可，无需安装 Go、Node.js 或解压发布包。Linux/macOS 下载后先执行 `chmod +x 文件名`，再执行 `./文件名`；Windows PowerShell 执行 `.\文件名.exe`，然后访问 [http://localhost:8080](http://localhost:8080)。所有前端资源直接从内存中的嵌入文件系统提供，不依赖外部 `web/dist` 目录，也不会在启动时解压到磁盘。

运行时数据库、附件、日志和密钥保存在 `DATA_DIR`（默认 `./data`），不会打进二进制。更新时停止旧进程、替换可执行文件，再启动即可，保留原有数据目录和环境变量配置。Supervisor 可直接启动该文件并通过 `environment=` 注入配置，原有 `WEB_DIST` 环境变量可删除。

## 环境变量

配置模板见 [.env.example](.env.example)。后端读取进程环境，**不会自动加载 `.env`**。本地需要配置时，可将模板复制为根目录 `.env`，修改后在启动后端前加载：

```bash
cp .env.example .env
# 编辑 .env 后执行：
set -a
. ./.env
set +a
./backend/bin/homework-station
```

包含空格或特殊字符的环境变量值需要按 shell 语法加引号。Docker Compose 会自动读取根目录 `.env`，无需手动加载。

| 变量             | 默认值                                      | 说明                                                       |
| ---------------- | ------------------------------------------- | ---------------------------------------------------------- |
| `PORT`           | `8080`                                      | HTTP 端口                                                  |
| `DATA_DIR`       | `./data`                                    | 数据库、附件、分片、日志和自动生成的会话密钥目录           |
| `DB_DRIVER`      | `sqlite`                                    | `sqlite`、`mysql`、`postgres`，兼容 `postgresql` / `pgsql` |
| `DB_DSN`         | SQLite 为 `DATA_DIR/homework.db`            | 完整数据库连接串，优先于分项配置                           |
| `DB_HOST`        | `localhost`                                 | MySQL / PostgreSQL 地址                                    |
| `DB_PORT`        | MySQL `3306`、PostgreSQL `5432`             | 数据库端口                                                 |
| `DB_USER`        | MySQL `root`、PostgreSQL `postgres`         | 数据库账号                                                 |
| `DB_PASSWORD`    | 空                                          | 数据库密码                                                 |
| `DB_NAME`        | `homework`                                  | 数据库名，限字母、数字和下划线                             |
| `DB_SSLMODE`     | `disable`                                   | PostgreSQL SSL 模式                                        |
| `JWT_SECRET`     | 首次生成并保存在 `DATA_DIR/.session-secret` | 手动配置时至少 32 字符；重启后须保持一致                   |
| `BINDING_SECRET` | 首次初始化时生成随机值                      | QQ 绑定密钥初始值；后续在后台修改                          |
| `COOKIE_SECURE`  | `false`                                     | HTTPS 部署设为 `true`                                      |
| `GIN_MODE`       | Gin 默认值                                  | 生产环境可设为 `release`                                   |

路径相对**后端进程的工作目录**。按上面的方式在项目根目录启动时，默认数据目录为 `./data/`；Supervisor 部署时以其 `directory=` 指定的目录为准。

### 数据库选择

未配置时直接使用 SQLite，无需独立数据库服务。

使用 MySQL 时设置 `DB_DRIVER=mysql`，使用 PostgreSQL 时设置 `DB_DRIVER=postgres`，并填写对应的地址、端口、账号、密码与库名。启动时后端尝试初始化目标库并执行表结构迁移，数据库账号需要相应的建库与迁移权限；也可提前由数据库管理员建库并授予运行账号所需权限。

建库通过 GORM 连接执行数据库对应的 `CREATE DATABASE` 语句；GORM 没有统一的建库接口。表、列、索引和外键则由 `AutoMigrate` 根据 Go 模型自动创建与迁移，无需手工导入 SQL。

作业支持独立的“作业简介”（可选，最多 200 字），在创建或编辑作业时填写，并展示在学生端作业卡片上；详情页仍使用完整作业描述。升级后首次启动会自动为 `assignments` 表新增 `summary` 列，已有作业的简介默认为空，不改写已有描述或提交记录；未填写时卡片显示“暂无作业简介”。升级前请按下文的备份说明备份数据库和附件。

完整 DSN 格式如下，均为占位示例，需替换为自己的连接信息：

```text
MySQL: user:password@tcp(database-host:3306)/homework?charset=utf8mb4&parseTime=True&loc=UTC
PostgreSQL: host=database-host port=5432 user=user password=password dbname=homework sslmode=disable
```

### Docker Compose

```bash
docker compose up --build -d
```

默认使用 SQLite，以非 root 用户运行，通过命名卷 `homework-data` 持久化 `/app/data`。

```bash
docker compose logs -f app
docker compose down
```

`docker compose down` 保留数据卷。连接外部数据库时通过 `.env` 配置连接信息；容器内 `localhost` 指容器自身。首次打开页面后仍需自行创建分类与作业、配置认证接口。

## 主要功能

### 用户前台

- 未登录访问 `/` 时显示登录页，可切换注册。
- 注册使用二次密码确认和可刷新的 4 位图形验证码，验证码 5 分钟有效且只能消费一次。
- “我的作业”支持名称搜索、多选分类、仅看未提交；筛选偏好保存到账号。
- 分类按树形层级缩进，作业按分类标题嵌套展示。选择父分类包含全部后代分类，未提交作业优先。
- 桌面端筛选面板位于右侧并粘性定位，窄屏移到列表上方。
- 作业详情提供 Markdown 说明、描述附件、文本/附件提交、历史记录与撤回。
- 个人信息页支持修改用户名、上传头像、凭旧密码修改密码，以及配置启用后的实名绑定和 QQ 绑定。

### 管理后台

`/admin` 提供总览、作业分类、作业管理、用户管理、附件管理、设置六个栏目。

- 总览：用户数、附件数、去重后的物理容量、历史提交数和最近 30 天的日增曲线。
- 分类：多级分类、排序、父级调整和删除；不限制固定层数，禁止循环引用，非空分类需先移动内容。
- 作业：名称、分类、发布状态、截止时间、排序、Markdown 编辑与预览、描述附件、允许的附件后缀/大小、多次提交及撤回规则。
- 历史提交：查看全部用户的历次提交，可删除记录，并选择清理没有其他用途的附件。
- 用户：编辑账号、重置密码、调整管理员权限、查看关注分类、提交历史、附件用途和绑定信息，支持解绑；至少保留一名管理员。
- 附件：搜索、分页、完成/未完成状态筛选、SHA256、归属、用途、下载和删除。
- 设置：站点标题、浏览器标签页标题、页面 Logo、站点 Icon、注册开关、默认上传限制、头像大小、实名接口配置、QQ 绑定密钥。

在 **设置 → 基础设置** 中可将“作业驿站”改为其他中文或英文名称（最多 120 个字符）。网站标题用于浏览器标签页，留空时跟随站点标题。Logo 显示在登录页和前后台导航栏，Icon 用于浏览器标签页。

Logo 与 Icon 支持填写 HTTP(S) 图片链接、站内绝对路径，或上传 PNG、JPEG、GIF、WebP 图片（最大 2MB、400 万像素，上传后转为 PNG 并保留透明背景）。上传完成后点击“保存设置”生效，也可清空地址或点击恢复默认。上传图片保存在 `DATA_DIR/branding/`，可公开访问，不占用用户私人附件库；备份与迁移时需一起保留该目录。

列表支持每页 10 / 20 / 50 / 100 条。后台账号菜单可进入当前用户编辑页或退出。

### 显示与默认规则

- 所有页面支持浅色、深色、跟随系统，默认跟随系统，主题偏好保存在当前浏览器。
- 主题色 `#FF406E`，深色主背景 `#0F1214`，卡片、侧栏、顶部导航背景 `#1B1D21`。
- 默认作业附件上限 20MB、后缀不限；头像上传上限 10MB。新作业读取当前默认值，已有作业保留各自规则。
- 截止时间留空代表长期可交。截止时间与提交限制同时由后端校验。
- 页面时间按浏览器时区显示，统计曲线按 UTC 日期归组。日容量曲线为当天新建且已完成附件的逻辑大小，总容量为去重后的物理大小。
- 统计基于当前保留的记录，删除内容后历史曲线可能随之变化。

## 实名认证配置

实名认证不包含预置服务地址，默认关闭。管理员进入 **设置 → 实名认证** 配置自己有权使用的接口：

| 配置项           | 默认值      | 用途                                           |
| ---------------- | ----------- | ---------------------------------------------- |
| 认证 API URL     | 空          | 留空关闭；配置后后端通过 GET 请求调用          |
| 成功判断字段     | `success`   | 从 JSON 响应提取成功状态                       |
| 成功字段的期望值 | `true`      | JSON 布尔值、数字、字符串或 null，严格区分类型 |
| 返回学号字段     | `data.id`   | 从响应提取学号，与用户输入比较                 |
| 返回姓名字段     | `data.name` | 从响应提取姓名，与用户输入比较                 |

URL 模板必须包含 `{id}`，可选 `{name}`。后端按路径或查询参数的位置自动编码，再替换为用户输入的学号和姓名。可使用如下**文档占位格式**理解配置方式，示例域名不提供实际服务：

```text
https://identity.example/verify?id={id}&name={name}
```

字段路径支持 `result.id`、`data.0.name`、`data[0].name` 和可选的 `$.` 前缀。例如成功字段为 `code` 时，期望值可填 `0`；期望字符串时需填写带双引号的 JSON 字符串，如 `"ok"`。

后台“站点设置 → 实名认证”可设置学号格式正则表达式和输入框占位提示（placeholder）。正则使用 Go（RE2）语法，不带 `/` 分隔符，自动匹配整个学号；默认 `^[0-9]{10}$`，例如 `^[A-Z]{2}[0-9]{8}$` 允许两位大写字母加八位数字。学号最多 128 个字符，非法正则会在保存时被拒绝。前台展示配置的提示，格式由后端统一校验，避免浏览器与服务端的正则语法差异。后端还验证成功状态，以及返回学号和姓名是否与输入一致。认证请求超时为 10 秒，不跟随重定向。未配置、网络错误、无效 JSON、字段缺失或身份不匹配均不会完成绑定。

配置保存后立即生效。API 地址及其参数只通过管理员接口读取，不出现在公开设置中；公开设置仅提供是否启用的状态。清空 URL 可关闭绑定入口，不影响已绑定用户的信息。一个学号只能绑定一个账号，更换身份需管理员先解绑。

## NoneBot QQ 绑定插件

插件位于 [nonebot_plugin_homework/](nonebot_plugin_homework/)，用于 NoneBot2 + OneBot v11。

1. 将插件目录放入机器人项目可导入的位置。
2. 安装依赖：

   ```bash
   pip install -r nonebot_plugin_homework/requirements.txt
   ```

3. 在机器人 `.env` 配置：

   ```dotenv
   COMMAND_START=["/"]
   HOMEWORK_BINDING_SECRET=与站点后台绑定码Secret完全相同的密钥
   ```

4. 在 NoneBot 启动文件加载：

   ```python
   nonebot.load_plugin("nonebot_plugin_homework")
   ```

用户私聊机器人发送 `/绑定码`，收到 JWT 后粘贴到个人信息页绑定。插件只响应私聊，绑定码 5 分钟有效且只能成功使用一次。

JWT 使用 HS256，包含字符串 `qq`、`ttl=300`、秒级 `iat` / `exp`、`aud=homework-qq-bind` 及随机 `jti`。后端校验签名、算法、用途、过期时间和 QQ 格式，同一 QQ 只能绑定一个账号。

密钥至少 32 字符；修改站点密钥后需同步机器人配置并重新加载插件。双方系统时钟需保持准确。密钥不要写入准备公开的代码或配置文件。

## 附件存储与部署范围

- 浏览器增量计算 SHA256，以 1MiB 分片上传；失败分片自动重试，支持暂停。
- 重新选择同一原文件可继续上传，服务端返回已收到的分片，无需从头传输。
- 合并时重新检查完整文件的大小与 SHA256。不同用户保留独立归属记录，相同内容只保存一份物理文件；其他用户不能仅凭哈希获取文件。
- 下载需要登录并验证权限；学生只能下载自己的提交附件以及可见作业的描述附件。
- 管理员下载未完成附件时得到包含清单和已接收分片的 ZIP，而不是损坏的“完整文件”。
- 头像由后端解码并转为 WebP，最多 2500 万像素；原上传大小由后台配置限制。
- 删除被引用附件需确认并解除关联；只有最后一个副本不再需要时才删除物理文件。描述中的手写 Markdown 链接需要管理员同步维护。

当前适合 **单后端进程 + 持久化本地目录** 部署，进程内锁协调上传、合并和删除；不能启动多个后端实例共享同一目录并发写入。单文件硬上限为 10GB，作业规则可进一步缩小限制。

认证使用 HttpOnly、SameSite=Strict Cookie 和 bcrypt 密码摘要；更改密码或管理员权限会使旧会话失效。HTTPS 部署时设置 `COOKIE_SECURE=true`，反向代理保留正确的 Host 与 Origin。

### 反向代理与真实 IP

`403 来源校验失败` 表示非 GET/HEAD 请求携带的 `Origin` 无法解析，或其中的域名、端口与后端收到的 `Host` 不一致。这是防止跨站请求的来源检查。常见原因是代理默认把 `Host` 改成 `127.0.0.1:8080` 或容器名，而浏览器的 `Origin` 仍是公网域名。应保留原始 Host（包含非标准端口），不要重写或删除浏览器的 Origin。

完整 Nginx 示例见 [docs/nginx.conf.example](docs/nginx.conf.example)。将域名替换为实际域名，在 Nginx 的 `http` 上下文中引入；HTTPS 证书与监听按实际部署配置。关键请求头如下：

```nginx
proxy_set_header Host $http_host;
proxy_set_header X-Real-IP $remote_addr;
proxy_set_header X-Forwarded-For $remote_addr;
proxy_set_header X-Forwarded-Proto $scheme;
```

示例适用于用户直接访问的单层 Nginx：覆盖用户传入的 IP 头。`$http_host` 会保留公网端口，避免使用 `$host` 丢失非标准端口。`X-Forwarded-Proto` 传递外部协议，Cookie 的 Secure 属性仍由 `COOKIE_SECURE` 配置控制。相关指令见 [Nginx 官方文档](https://nginx.org/en/docs/http/ngx_http_proxy_module.html#proxy_set_header)。

后端通过 `TRUSTED_PROXIES` 指定可信代理，支持逗号或空白分隔的 IPv4、IPv6 及 CIDR。例如，Nginx 与原生后端在同一主机，通过回环地址连接时：

```dotenv
TRUSTED_PROXIES=127.0.0.1,::1
COOKIE_SECURE=true
```

`COOKIE_SECURE=true` 适用于 HTTPS；本地 HTTP 测试保留 `false`。原生运行时需先导出这些环境变量；Docker Compose 会从 `.env` 传给容器。修改后重启后端，Compose 部署执行 `docker compose up -d --build`。

仅在连接后端的直接对端属于可信代理时，后端才读取 `X-Forwarded-For`，其次读取 `X-Real-IP`；日志和登录/注册限流使用识别出的 IP。多层代理的 `X-Forwarded-For` 从右向左检查，选择第一个不可信地址；需要在入口覆盖外部传入的 IP 头，内部代理再追加地址，并配置实际可信代理链。机制见 [Gin 官方说明](https://github.com/gin-gonic/gin/blob/master/docs/doc.md#dont-trust-all-proxies)。

未设置 `TRUSTED_PROXIES` 时默认信任 `127.0.0.1`；显式设置 `TRUSTED_PROXIES=` 则不信任任何转发 IP 头，原生运行与 Docker Compose 行为一致。已有 `.env` 中若保留空值，需要删除该行或改为 `TRUSTED_PROXIES=127.0.0.1` 才会使用回环代理。无效的 IP/CIDR 会导致启动失败。不要配置 `0.0.0.0/0` 或 `::/0`；只填写实际代理地址或受控网络。Docker 中直接对端可能是网桥网关或代理容器 IP，并非 `127.0.0.1`，应以实际连接地址配置。宿主机 Nginx 代理 Docker 后端时，建议把 Compose 端口映射改为 `127.0.0.1:8080:8080`；同一容器网络内代理时使用服务名 `http://app:8080`，并限制后端只能由代理访问。

备份需要同时考虑数据库和附件：SQLite 部署可在停服后备份整个 `DATA_DIR`；外部数据库需另做与文件目录一致的备份。迁移时保留会话密钥，修改 `DATA_DIR` 前同时迁移已有文件和数据库中的存储路径。

## 测试与仓库内容

```bash
make test
```

- Go 测试使用 `t.TempDir()` 创建隔离数据库和附件目录，结束后自动清理；实名认证使用模拟网络响应，不请求真实认证服务。
- 前端测试验证分类树、移动规则、嵌套标题和筛选分组；Python 测试验证绑定 JWT 合约。
- 前端生产构建同时进行 TypeScript 检查。
- 初始化回归测试明确验证：仅有一个默认管理员，分类、作业、提交、附件等业务表均为空，且未配置认证 API 时不发起外部请求。

仓库不提供演示数据库、预填业务数据、真实认证 API、运行密钥或带账号/配置的调试截图。自动化测试中用于断言的临时合成输入不会写入开发数据库、生产数据库或生产构建产物。

`.gitignore` 与 `.dockerignore` 排除本地 `.env`、数据库、运行目录、日志、密钥文件、调试截图和生成的后端可执行文件。配置模板中的地址和连接信息均为占位格式。

验证方式与范围见 [docs/verification.md](docs/verification.md)。

## 工程结构

前后端均按模块拆分，入口只负责装配。详细目录、依赖方向和新增模块约定见 [工程架构](docs/architecture.md)。

```text
backend/
  cmd/server/             可执行程序入口、HTTP 生命周期
  internal/
    app/                  依赖装配、资源关闭、隔离集成测试
    config/               环境变量配置
    database/             各数据库连接、迁移、管理员初始化
    http/
      router/             API 路由、SPA 回退
      middleware/         鉴权、管理员权限、频率及来源校验
      handler/            按资源划分的请求处理器
      response/           HTTP 错误响应
    model/                GORM 数据模型
    security/             密码与随机值工具
    service/              业务规则、事务、认证服务、附件存储
web/src/
  app/                    路由、守卫、全局 providers
  components/             通用表单、布局、反馈、shadcn/ui
  features/               按业务拆分的 pages、components、hooks、api、lib
  hooks/                  通用数据加载 hook
  lib/                    HTTP 客户端、URL、格式化工具
  styles/                 主题、基础、布局、业务样式与响应式覆盖
  types/                  API 数据类型
nonebot_plugin_homework/   QQ 绑定插件
scripts/                  工程维护脚本
tests/                    Python JWT 合约测试
```

## 统一格式化

开发环境额外需要 Python 3.10+。在根目录安装开发工具：

```bash
make install-dev
```

执行全仓库格式化或只检查差异：

```bash
make format
make format-check
```

| 范围                                                         | 工具/配置                                                                                 |
| ------------------------------------------------------------ | ----------------------------------------------------------------------------------------- |
| Go                                                           | `goimports`，版本通过 `backend/go.mod` 的 tool 指令锁定，同时整理导入与执行 Go 标准格式化 |
| TypeScript、TSX、JavaScript、CSS、HTML、JSON、YAML、Markdown | Prettier，根目录 `.prettierrc.json` / `.prettierignore`                                   |
| Python                                                       | Ruff 格式化与导入排序，`pyproject.toml` / `requirements-dev.txt`                          |
| TOML                                                         | Taplo，`.taplo.toml`                                                                      |
| Makefile、Dockerfile、环境模板与忽略文件                     | `.editorconfig` 与根格式化脚本的行尾/空白规范化，保留 Makefile 的 Tab                     |

`make format-check` 不会修改文件，存在差异时以非零状态退出。也可以进入 `web/` 单独运行 `npm run format` 或 `npm run format:check`。

`.vscode/` 提供保存时格式化设置及推荐扩展；GitHub Actions 会执行格式检查、测试、生产构建和 `go vet`，并使用独立的 MySQL 8.0 服务测试初始化、配置读写与重启后的配置保留。依赖、运行数据、日志、密钥与构建产物均排除在格式化范围外。

本地默认测试使用临时 SQLite 数据库；若要执行 MySQL 回归测试，在运行 `make test` 前设置 `MYSQL_TEST_DSN`，指向可丢弃的专用测试库（例如 `root:test-password@tcp(127.0.0.1:3306)/homework_ci?charset=utf8mb4&parseTime=True&loc=UTC`）。该测试会初始化表、管理员和配置并修改配置，不要使用生产数据库；未设置时跳过此测试。
