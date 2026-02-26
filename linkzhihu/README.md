# 知乎用户回答爬虫

这是一个用于爬取知乎用户所有回答的Python爬虫工具。

## 功能特点

✅ **自动登录保持**：首次扫码登录后，登录状态会自动保存，下次运行无需重复登录  
✅ **断点续传**：支持中断后继续爬取，不会重复抓取已有数据  
✅ **反爬虫策略**：随机延迟、模拟真实浏览器行为  
✅ **数据导出**：自动保存为CSV格式，方便后续分析  

## 快速开始

### 1. 安装依赖

```bash
pip install DrissionPage pandas
```

### 2. 配置目标用户

打开 `config.py`，修改 `TARGET_USER_ID`：

```python
# 从知乎用户主页URL中提取用户ID
# 例如：https://www.zhihu.com/people/dong-bu-dong-95-73
TARGET_USER_ID = 'dong-bu-dong-95-73'
```

### 3. 运行爬虫

```bash
python spider.py
```

### 4. 首次登录

首次运行时：
1. 程序会自动打开Chrome浏览器
2. 浏览器会跳转到知乎首页
3. **请在60秒内完成扫码登录**（可在config.py中修改等待时间）
4. 登录成功后，程序会自动开始爬取

### 5. 查看结果

爬取的数据会保存在 `output` 目录下，文件名格式：
```
{用户ID}_answers_{时间戳}.csv
```

## 配置说明

### 基础配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `TARGET_USER_ID` | 目标用户ID（从URL中提取） | `'dong-bu-dong-95-73'` |
| `MAX_ANSWER_COUNT` | 最大爬取数量（0=不限制） | `0` |

### 登录配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `LOGIN_WAIT_TIME` | 登录等待时间（秒） | `60` |

### 反爬虫配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `RANDOM_SLEEP_MIN` | 最小随机等待时间（秒） | `3` |
| `RANDOM_SLEEP_MAX` | 最大随机等待时间（秒） | `6` |

### 重试配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `MAX_RETRY` | 最大重试次数 | `3` |
| `RETRY_INTERVAL` | 重试间隔（秒） | `5` |

## 数据字段说明

导出的CSV文件包含以下字段：

| 字段 | 说明 |
|------|------|
| Question Title | 问题标题 |
| Answer ID | 回答ID |
| Excerpt | 回答摘要 |
| Content | 回答完整内容（HTML格式） |
| Vote Count | 点赞数 |
| Comment Count | 评论数 |
| Create Time | 创建时间 |
| URL | 回答链接 |

## 常见问题

### Q: 如何获取用户ID？

A: 打开知乎用户主页，URL格式为：
```
https://www.zhihu.com/people/用户ID
```
例如：`https://www.zhihu.com/people/dong-bu-dong-95-73`，用户ID就是 `dong-bu-dong-95-73`

### Q: 不需要Token吗？

A: **不需要**。本脚本使用浏览器会话登录，首次扫码登录后，登录信息会保存在 `browser_data` 目录中，下次运行会自动使用已保存的登录状态。

### Q: 如何限制爬取数量？

A: 修改 `config.py` 中的 `MAX_ANSWER_COUNT`：
```python
MAX_ANSWER_COUNT = 100  # 只爬取100条回答
```

### Q: 遇到验证码怎么办？

A: 程序会自动暂停，提示您手动处理验证码。处理完成后按回车继续。

### Q: 如何清除登录状态重新登录？

A: 删除 `browser_data` 目录，下次运行时会要求重新登录。

### Q: 程序中断后如何继续？

A: 直接再次运行 `python spider.py`，程序会自动从上次中断的位置继续爬取（通过 `progress.json` 记录进度）。

## 目录结构

```
linkzhihu/
├── spider.py           # 主爬虫程序
├── config.py           # 配置文件
├── test_spider.py      # 单元测试
├── browser_data/       # 浏览器数据目录（登录状态）
├── output/             # 输出目录（CSV文件）
├── progress.json       # 进度文件（断点续传）
└── spider.log          # 运行日志
```

## 运行测试

```bash
python test_spider.py
```

## 注意事项

⚠️ **请遵守知乎的使用条款和robots.txt规则**  
⚠️ **建议设置合理的爬取间隔，避免对服务器造成压力**  
⚠️ **仅用于学习和个人研究，请勿用于商业用途**  

## License

MIT
