<h1 align="center">Go_Frontend</h1>
<p align="center">一个实现 Emby 服务播放前后端分离的程序套件。</p>

## 简介

1. 本项目是实现 Emby 媒体服务播放前后端分离的前端程序，需要与播放分离后端 [Go Playback Backend](https://github.com/Moxi007/Go_Backend) 配套使用。
2. 本程序很大程度上基于 [YASS-Frontend](https://github.com/FacMata/YASS-Frontend)。原版是用 `Python` 实现的，为了获得更好的兼容性，已重写为 `Go` 版本并在其基础上进行了优化，使其更加易用。

------

## 原理

1. 使用特定的 `nginx` 配置（参考 [nginx.conf](https://github.com/Moxi007/Go_Frontend/blob/main/nginx/nginx.conf)）将 Emby 播放链接重定向到指定端口。
2. 程序监听该端口接收到的请求，并提取 `MediaSourceId` 和 `ItemId`。
3. 向 Emby 服务请求对应的文件相对路径（`EmbyPath`）。
4. **确定后端**：将 `EmbyPath` 与配置的 `Backends` 列表进行匹配（最长前缀匹配），以选择合适的流媒体服务器并生成相对路径。
5. 通过将配置中的 `Encipher` 值与过期时间 (`expireAt`) 进行加密来生成签名 `signature`。
6. 将后端播放地址 (`backendURL`) 与匹配到的相对路径和 `signature` 进行拼接。
7. 将播放请求重定向到生成的 URL，交由后端处理。

------

## 功能

- **兼容所有版本的 Emby 服务器**。
- **多后端支持**：配置多个存储后端，基于文件路径（最长前缀匹配）进行智能路由。
- **高性能**：
    - **Singleflight**：防止热点视频的缓存击穿（惊群效应），保护 Emby 服务器。
    - **HTTP Keep-Alive**：复用与 Emby API 的 TCP 连接，降低延迟并减少端口占用。
- **支持高并发**，可同时处理多个请求。
- **支持部署了 `strm` 的 Emby 服务器**。
- **请求缓存**，对相同的 `MediaSourceId` 和 `ItemId` 请求进行快速响应，减少起播时间。
- **链接签名**，由前端生成签名，后端验证签名。签名不匹配将导致 `401 Unauthorized` 错误。
- **链接过期**，签名中嵌入了过期时间，防止恶意抓包导致链接被长期盗用。

------

## 配置文件

```yaml
# Logging configuration
LogLevel: "INFO" # 日志级别 (例如: info, debug, warn, error)

# Encryption settings
Encipher: "vPQC5LWCN2CW2opz" # 用于加密和混淆的密钥

# Emby server configuration
Emby:
  url: "[http://127.0.0.1](http://127.0.0.1)" # Emby 服务器的基础 URL
  port: 8096
  apiKey: "6a15d65893024675ba89ffee165f8f1c"  # 用于访问 Emby 服务器的 API 密钥

# 多后端配置 (Multiple Backend Configuration)
# 程序会将 Emby 文件路径与每个后端的 'path' 进行匹配。
# 它会自动优先匹配最长的路径前缀。
Backends:
  - name: "Anime Drive"
    url: "[https://stream-anime.example.com/stream](https://stream-anime.example.com/stream)"  # 该后端的公开流媒体 URL
    path: "/mnt/anime"                               # Emby 中的绝对路径前缀

  - name: "Movie Drive"
    url: "[https://stream-movie.example.com/stream](https://stream-movie.example.com/stream)"
    path: "/mnt/movies"

  - name: "General Storage"
    url: "[https://stream-general.example.com/stream](https://stream-general.example.com/stream)"
    path: "/mnt/share"

# Streaming configuration
PlayURLMaxAliveTime: 21600 # 播放链接的最大存活时间，单位秒 (例如: 6 小时)

# Server configuration
Server:
  port: 60001

# Special medias configuration
SpecialMedias:
   # 下面的键值可以根据需要填写。如果不需要，可以留空。
   - key: "MediaMissing"
     name: "Default media for missing cases"
     mediaPath: "specialMedia/mediaMissing"
     itemId: "mediaMissing-item-id"
     mediaSourceID: "mediaMissing-media-source-id"
   - key: "September18"
     name: "September 18 - Commemorative Media"
     mediaPath: "specialMedia/september18"
     itemId: "september18-item-id"
     mediaSourceID: "september18-media-source-id"
   - key: "October1"
     name: "October 1 - National Day Media"
     mediaPath: "specialMedia/october1"
     itemId: "october1-item-id"
     mediaSourceID: "october1-media-source-id"
   - key: "December13"
     name: "December 13 - Nanjing Massacre Commemoration"
     mediaPath: "specialMedia/december13"
     itemId: "december13-item-id"
     mediaSourceID: "december13-media-source-id"
   - key: "ChineseNewYearEve"
     name: "Chinese New Year's Eve Media"
     mediaPath: "specialMedia/chinesenewyeareve"
     itemId: "chinesenewyeareve-item-id"
     mediaSourceID: "chinesenewyeareve-media-source-id"
```
------

## 如何使用

### 1. Docker 安装 (推荐)

#### 1.1 创建目录

```shell
mkdir -p /data/docker/go_frontend
```

#### 1.2 创建配置文件

```shell
cd /data/docker/go_frontend
mkdir -p config && cd config
```

将 config.yaml 复制到 config 文件夹中，并根据实际情况编辑。

#### 创建 docker-compose.yaml

返回 /data/docker/pilipili_backend 目录，将 docker-compose.yml 复制到该目录下。

#### 1.4 启动容器

```shell
docker-compose pull && docker-compose up -d
```

### 2. 手动安装

#### 2.1 安装 Go 环境

##### 2.1.1 卸载旧版本 (可选)

```bash
rm -rf /usr/local/go
```

##### 2.1.2 下载并安装

```bash
wget -q -O /tmp/go.tar.gz https://go.dev/dl/go1.23.5.linux-amd64.tar.gz && tar -C /usr/local -xzf /tmp/go.tar.gz && rm /tmp/go.tar.gz
```

##### 2.1.3 配置环境变量

```bash
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc && source ~/.bashrc
```

##### 2.1.4 验证安装

```bash
go version 
# 预期输出类似: go version go1.23.5 linux/amd64
```

------

#### 2.2 下载代码

```bash
git clone [https://github.com/Moxi007/Go_Frontend/.git](https://github.com/Moxi007/Go_Frontend/.git) /data/emby_backend
```

------

#### 2.3 编译与配置

进入目录并编辑`config.yaml`：

```shell
cd /data/emby_backend
vi config.yaml
```
编译二进制文件（推荐，比 go run 性能更好）：

```shell
go build -ldflags="-s -w" -o pilipili_backend main.go
```

#### 2.4 运行程序

```shell
# 前台运行测试
./pilipili_backend config.yaml

# 后台运行
nohup ./pilipili_backend config.yaml > streamer.log 2>&1 &
```
