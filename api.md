# 下载任务管理API文档

## 基础信息
- 协议：HTTP
- 数据格式：JSON
- 基础路径：`/api/tasks`


## 1. 增加任务（创建下载任务）

### 概述
创建一个新的下载任务，支持设置HTTP头、下载参数等配置

### 请求信息
- **方法**：POST
- **URL**：`/api/tasks`
- **请求体**：
```json
  {
    "url": "string",               // 必选，下载链接（M3U8地址）
    "filename": "string",          // 必选，保存的文件名
    "use_http_headers": boolean,   // 可选，是否使用自定义HTTP头，默认false
    "http_headers": {              // 可选，自定义HTTP头，当use_http_headers为true时有效
      "Header-Key1": ["Value1"],
      "Header-Key2": ["Value2"]
    },
    "advanced": boolean,           // 可选，是否使用高级设置，默认false
    "save_location": "string",     // 可选，文件保存路径（高级设置或系统默认）
    "download_threads": number,    // 可选，下载线程数
    "skip_ts_info": "string",      // 可选，跳过的TS片段表达式（如"1,92-100"）
    "keep_ts_files": boolean,      // 可选，是否保留TS文件
    "no_merge_ts": boolean,        // 可选，是否不合并TS为MP4
    "log_skipped_ts": boolean,     // 可选，是否记录跳过的TS信息
    "allow_insecure_https": boolean, // 可选，是否允许不安全的HTTPS请求
    "proxy_type": "string",        // 可选，代理类型
    "proxy_host": "string",        // 可选，代理主机
    "proxy_port": number,          // 可选，代理端口
    "proxy_username": "string",    // 可选，代理用户名（如需认证）
    "proxy_password": "string",    // 可选，代理密码（如需认证）
    "use_server_file_time": boolean, // 可选，是否使用服务端文件时间
    "debug_log": boolean           // 可选，是否开启调试日志
  }
  ```

### 响应信息
- **成功响应**（200 OK）：
```json
  {
    "status": "success",
    "id": "任务唯一ID"  // 新创建任务的ID
  }
  ```
- **错误响应**（400 Bad Request / 500 Internal Server Error）：
```json
  "错误描述信息"  // 如"解析请求失败"、"保存任务失败"
  ```


## 2. 查询所有任务

### 概述
获取系统中所有下载任务的详细信息，包括状态、进度等

### 请求信息
- **方法**：GET
- **URL**：`/api/tasks`
- **请求参数**：无

### 响应信息
- **成功响应**（200 OK）：
```json
  [
    {
      "id": "string",               // 任务ID
      "url": "string",              // 下载链接
      "filename": "string",         // 文件名
      "size": number,               // 文件大小（字节）
      "progress": number,           // 下载进度（百分比，0-100）
      "status": "string",           // 任务状态（"等待中"、"下载中"、"已暂停"、"已完成"、"失败"）
      "status_bar": "string",       // 状态详情文本
      "err_msg": "string",          // 错误信息（状态为"失败"时可能存在）
      "created_at": "string",       // 创建时间（ISO格式）
      "updated_at": "string",       // 更新时间（ISO格式）
      "completed_at": "string",     // 完成时间（ISO格式，状态为"已完成"时存在）
      "header_map": {               // 自定义HTTP头
        "Header-Key": ["Value"]
      },
      "advanced_settings": {        // 高级设置信息
        "save_location": "string",
        "download_threads": number,
        // 其他高级设置字段...
      }
    },
    // 更多任务...
  ]
  ```


## 3. 更新任务状态

### 概述
更新指定任务的状态（如暂停、继续等）

### 请求信息
- **方法**：PUT
- **URL**：`/api/tasks?id=任务ID`  // 任务ID通过查询参数传递
- **请求体**：
```json
  {
    "status": "string"  // 必选，目标状态，可选值："等待中"、"已暂停"
  }
  ```

### 响应信息
- **成功响应**（200 OK）：
```json
  {
    "status": "success"
  }
  ```
- **错误响应**：
- 400 Bad Request：`"任务ID不能为空"`
- 404 Not Found：`"任务不存在"`
- 500 Internal Server Error：`"更新任务失败"`


## 4. 删除任务

### 概述
删除指定的下载任务

### 请求信息
- **方法**：DELETE
- **URL**：`/api/tasks?id=任务ID`  // 任务ID通过查询参数传递
- **请求参数**：无请求体

### 响应信息
- **成功响应**（200 OK）：
```json
  {
    "status": "success"
  }
  ```
- **错误响应**：
- 400 Bad Request：`"任务ID不能为空"`
- 500 Internal Server Error：`"删除任务失败"`

## 5. 进度推送WebSocket（/ws/progress）

### 概述
该WebSocket端点用于实时推送下载任务的进度更新和状态详情，专注于高效传递任务进度相关信息，适用于需要快速获取任务动态的场景。

### 连接地址
- 路径：`/ws/progress`
- 协议：WebSocket (ws) 或 加密WebSocket (wss，取决于服务端是否启用HTTPS)


### 推送格式
推送数据为JSON格式，结构基于`FastPushData`结构体，包含以下字段：

| 字段名      | 类型   | 说明                     |
|-------------|--------|--------------------------|
| `id`        | string | 任务唯一ID（可选，关联具体任务） |
| `status_bar`| string | 任务当前状态描述（如"正在下载TS片段"） |
| `progress`  | int    | 下载进度（百分比，0-100） |


#### 示例
```json
{
  "id": "Ux6R7eWl4L_2",
  "status_bar": "[3/5]下载ts 速度 5.09 MB/s, 剩余时间 00:07",
  "progress": 69
}
```

```json
{
  "id": "Ux6R7eWl4L_2",
  "status_bar": "[1/5]嗅探m3u8 "
}
```


### 说明
- 当`id`字段存在时，表示该消息对应特定任务的进度更新
- 当任务处于初始化、准备等阶段时，`progress`可能为0且`id`可能暂未生成
- 推送频率根据任务进度变化动态调整，通常在进度发生变化时立即推送

## 状态码说明
- 200 OK：请求成功
- 400 Bad Request：请求参数错误或格式不正确
- 404 Not Found：请求的任务不存在
- 405 Method Not Allowed：使用了不支持的HTTP方法
- 500 Internal Server Error：服务器内部错误