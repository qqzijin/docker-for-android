# docker-for-android

目前暂时支持 arm64/aarch64 架构的部分 Android 设备。

## 版本与发布

- 当前版本：`28.0.1.10`
- 源服务器（原始）：`https://fw.koolcenter.com/binary/docker-for-android`
- CDN（.txt 不缓存，始终最新）：`https://fw.kspeeder.com/binary/docker-for-android`

### 打包 arm64 二进制

在仓库根目录执行：

```
make arm64
```

会将 `docker/arm64_bin` 打包为 `release/docker-for-android-bin-28.0.1.10-arm64.tar.gz`，并生成对应的 `.sha256` 校验文件。

生成 `version.txt`：

```
make version
```

或者一键构建：

```
make build-release
```

`version.txt` 中会包含版本号与 arm64/x86_64 包名与 sha256，用于脚本在 CDN 上获取并校验下载。

## 从 CDN 更新 arm64_bin

开发调试时，可以使用脚本从 CDN 获取 `version.txt`，并根据其中的信息下载并覆盖 `docker/arm64_bin`：

```
scripts/update_arm64_bin.sh
```

可通过环境变量覆盖默认端点：

```
CDN_URL=https://fw.kspeeder.com/binary/docker-for-android \
ORIGIN_SERVER_URL=https://fw.koolcenter.com/binary/docker-for-android \
scripts/update_arm64_bin.sh
```

脚本会：

- 从 CDN 下载 `version.txt`（.txt 不被 CDN 缓存，始终最新）
- 解析 `ARM64_PACKAGE` 与 `ARM64_SHA256`
- 从 CDN 下载 tar.gz（失败则回退到源服务器）
- 校验 sha256
- 解压到 `docker/` 下，覆盖 `docker/arm64_bin`

