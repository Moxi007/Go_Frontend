<h1 align="center">Go_Frontend</h1>
<p align="center">A program suite for separating the frontend and backend of Emby service playback.</p>

![Commit Activity](https://img.shields.io/github/commit-activity/m/hsuyelin/PiliPili_Frontend/main) ![Top Language](https://img.shields.io/github/languages/top/hsuyelin/PiliPili_Frontend) ![Github License](https://img.shields.io/github/license/hsuyelin/PiliPili_Frontend)


[中文版本](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/README_CN.md)

## Introduction

1. This project is a frontend application for separating Emby media service playback into frontend and backend components. It works in conjunction with the playback backend [PiliPili Playback Backend](https://github.com/hsuyelin/PiliPili_Backend).
2. This program is largely based on [YASS-Frontend](https://github.com/FacMata/YASS-Frontend). The original version was implemented in `Python`. To achieve better compatibility, it has been rewritten in `Go` and optimized to enhance usability.

------

## Principles

1. Use a specific `nginx` configuration (refer to [nginx.conf](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/nginx/nginx.conf)) to redirect Emby playback URLs to a designated port.
2. The program listens for requests arriving at the port and extracts the `MediaSourceId` and `ItemId`.
3. Request the corresponding file's relative path (`EmbyPath`) from the Emby service.
4. **Determine Backend**: Match the `EmbyPath` against the configured `Backends` list (longest prefix match) to select the appropriate streaming server and generate the relative path.
5. Generate a signed URL by encrypting the configuration's `Encipher` value with the expiration time (`expireAt`) to create a `signature`.
6. Concatenate the backend playback URL (`backendURL`) with the matched relative path and `signature`.
7. Redirect the playback request to the generated URL for backend handling.

![sequenceDiagram](https://github.com/hsuyelin/PiliPili_Frontend/blob/main/img/sequenceDiagram.png)

------

## Features

- **Compatible with all Emby server versions**.
- **Multi-Backend Support**: Configure multiple storage backends with intelligent routing based on file paths (longest prefix matching).
- **High Performance**:
    - **Singleflight**: Prevents cache stampede (thundering herd problem) for hot videos, protecting the Emby server.
    - **HTTP Keep-Alive**: Reuses TCP connections to the Emby API, reducing latency and port usage.
- **Supports high concurrency**, handling multiple requests simultaneously.
- **Support for deploying Emby server with `strm`.**
- **Request caching**, enabling fast responses for identical `MediaSourceId` and `ItemId` requests, reducing playback startup time.
- **Link signing**, where the frontend generates and the backend verifies the signature. Mismatched signatures result in a `401 Unauthorized` error.
- **Link expiration**, with an expiration time embedded in the signature to prevent unauthorized usage and continuous theft via packet sniffing.

------

## Configuration File

```yaml
# Logging configuration
LogLevel: "INFO" # Log level (e.g., info, debug, warn, error)

# Encryption settings
Encipher: "vPQC5LWCN2CW2opz" # Key used for encryption and obfuscation

# Emby server configuration
Emby:
  url: "[http://127.0.0.1](http://127.0.0.1)" # The base URL for the Emby server
  port: 8096
  apiKey: "6a15d65893024675ba89ffee165f8f1c"  # API key for accessing the Emby server

# Multiple Backend Configuration
# The program will match the Emby file path against the 'path' of each backend.
# It automatically prioritizes the longest matching path prefix.
Backends:
  - name: "Anime Drive"
    url: "[https://stream-anime.example.com/stream](https://stream-anime.example.com/stream)"  # The public streaming URL for this backend
    path: "/mnt/anime"                               # The absolute path prefix in Emby

  - name: "Movie Drive"
    url: "[https://stream-movie.example.com/stream](https://stream-movie.example.com/stream)"
    path: "/mnt/movies"

  - name: "General Storage"
    url: "[https://stream-general.example.com/stream](https://stream-general.example.com/stream)"
    path: "/mnt/share"

# Streaming configuration
PlayURLMaxAliveTime: 21600 # Maximum lifetime of the play URL in seconds (e.g., 6 hours)

# Server configuration
Server:
  port: 60001

# Special medias configuration
SpecialMedias:
   # The key values below can be filled as needed. If not required, they can be left empty.
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
