# 框架优化规划

各模块待优化项与生产就绪补充项。

---

## 待办

### SSE 背压控制（暂缓）

**位置:** `sse/util/sse_writer.go` — 新增 `WriteSSEDrop`

**问题:** 生产速度 > 消费速度时，TCP 发送缓冲区填满导致 `Write()` 阻塞或 OOM。

**方案:** 提供非阻塞写入（channel + goroutine 异步模型），写入失败丢弃当前帧，业务层自行决定丢弃或降速策略。

---

## 框架优化点与风险

| # | 位置 | 问题 | 建议 |
|---|------|------|------|
| 1 | `framework/service/ioc` | `Container.values` 无锁，运行时并发 `Set` 有数据竞争 | ✅ 已约定：`docs/ioc_injection.md`「并发约定」+ `ioc_svc.go` 注释明确 Set 仅初始化阶段调用 |
| 2 | `framework/service/http` | `registerAPI` 每请求 `reflect.New` + `Inject` 反射开销 | ✅ 已是最优：`apiType` 已在闭包外缓存，`reflect.New` 每请求不可避免 |
| 3 | `sse` | `sseWriter` 用 mutex 串行化心跳与业务写入，高频推送有锁竞争 | 4.1 异步模型可缓解 |
| 4 | `log-sls` | `sls_svc` 中 `_ = client.PutLogs(...)` 忽略错误，上传失败静默丢失 | 暂缓（决定不管） |
| 5 | `framework/service/http` | API 层无请求频率限流（防刷） | ✅ 已实现：限流拆为独立 `ratelimit` 模块（内存 + Redis），按 IP 限流，超限 429 |
| 6 | 各模块 | 各模块已有基础单测，framework 覆盖偏浅（仅构造/util） | 视需要补 error 契约、response 枚举/响应码映射 |
| 7 | `docs/` | 文档全部中文，英文用户仅有 README_EN.md | 后续翻译 docs |
| 8 | `example/config.yaml` | SLS 密钥明文（用户已知，自行处理） | 轮换 + 环境变量注入 |

---

## 模块拆分状态

已完成：框架拆为 8 个独立仓库，挂在 GitHub 组织 `hecc-blot` 下（`github.com/hecc-blot/{framework,ratelimit,log-sls,sse,db,cache,trace}` + 伞仓 `guide`）。当前暂不打 tag，等框架架构稳定后再发布版本。

**依赖拓扑**（推送/升级顺序）：`framework`/`ratelimit` 无内部依赖；`log-sls`/`sse`/`db`/`trace → framework`；`cache → framework+trace`；`guide → 其余 7 个`。
