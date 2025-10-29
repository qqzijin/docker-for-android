这个项目是为原生的 Android 系统提供标准的 Docker 能力。
目前的原理是：
把 bin 目录的文件打包到服务器的目录：
https://fw.koolcenter.com/binary/docker-for-android
服务器对应的 CDN 地址为：
https://fw.kspeeder.com/binary/docker-for-android
CDN 地址中 .txt 的后缀的文件是不会被缓存的。

用户可以通过 adb push 脚本到服务器上，然后下载从服务器下载文件，放到正确的位置上，然后启动 docker。
如果目录当中已经有 Docker，则重启的时候，也能保证 Docker 自动被启动。

为了支持 arm64 跟 x86_64 两个平台，以及为了支持 CDN，我们的文件名字最好带上版本号。
版本号我们跟随 docker 版本以及加上我们的子版本号。
目前我们已经有了一个设置版本号的文件：VERSION
从 VERSION 里面可以读到我们真正的版本号。为了方便沟通，加入从里面读到的版本号为：28.0.1.10
我们打包的时候需要打包为：

* docker-for-android-bin-28.0.1.10-arm64.tar.gz
* docker-for-android-bin-28.0.1.10-x86_64.tar.gz

咱们把 CDN 地址作为一个变量，然后在脚本中使用。
同时把原服务器也作为一个变量放到脚本中使用。
我们应该设置一个 version.txt 文件，这个文件是不会缓存到 CDN 当中的。
把上面两个文件的 sha256 放到 version.txt 中，同时把 28.0.1.10 的版本号也放到 version.txt 文件中。

已经实现一个 Makefile，Makefile 里面实现打包 bin 目录的功能，并把打包得到的两个文件放到一个 release 目录中。
实现 Makefile 的过程中呢，我们只需要考虑 arm64 的情况即可，arm64 的 bin 目录就是 arm64_bin

make arm64 应该能打包出来文件已经放到 release 中

上面的二进制已经打包完成，我们还需要打包 docker 文件夹，变成 docker-28.0.1.10.tar.gz
docker 文件夹非常小，是 arm64 x86_64 公用的。打包的时候要排除掉 arm64_bin x86_64_bin.
目前已经进一步优化好了 Makefile，实现 docker 的打包，并正好到之前的打包的 release 目录中。

我们已经完成了， docker 文件夹实现一个脚本，当 docker 文件夹被下载到了 /data/local/docker 之后，运行 /data/local/docker/deploy-in-android.sh 就会自动初始化好 docker 环境。
初始化 docker 环境的步骤是：

1. 判断是否接入硬盘，且硬盘格式化并挂载好了。参考下面的代码，是可以运行在 Android adb shell 的 root 环境中的：

	NVME=$(mount | grep -F '/dev/block/vold/public:259,1 on /mnt/media_rw/' | grep -F ' type ext4 ' | grep -oE '/mnt/media_rw/[^ ]+')
	if [ -n "$NVME" ]; then
		DOCKER_DATA_ROOT="$NVME/opt/dockerd/docker"
		mkdir -p "$DOCKER_DATA_ROOT"
		touch "$NVME/opt/.nomedia"
	fi

2. 得到了硬盘的挂载点后，设置 docker 目录下面的 docker.env 文件，把这个变量设置为具体的值：
export DOCKER_DATA_ROOT=

3. 同时修改 docker 目录下面的 etc/kspeeder.conf 把下面的 -cachePath 目录设置为上面得到的挂载点：
command=/data/local/docker/bin/kspeeder \
    -cachePath /mnt/media_rw/048a3593-e231-438c-a980-38149429f948/dockercache

如果 supervisor 也能自动从 docker.env 里面得到环境变量，并传递给 kspeeder 就更好了。
这样可以更好的统一环境变量。 

4. 之后需要启动 docker，检测 docker 是否启动完成。
启动完成后下载并安装一个 dpanel 的容器，监听在 :8080 端口上，且 dpanel 的容器的配置数据目录映射到
DOCKER_DATA_ROOT 的上层文件夹的 DOCKER_DATA_ROOT/../Configs/DPanel 目录下

经过思考，我们还是需要更简单些。
默认我们找到一个合适的硬盘位置，定义为 DISK_ROOT
后续的 DOCKER_DATA_ROOT KSPEEDER 路径都从 DISK_ROOT 派生出来。
这样我们只需要找到一个硬盘，并设置硬盘的的变量，后续的其他软件都延续这个硬盘位置。比如下面这样：
export DISK_ROOT=
export DOCKER_DATA_ROOT=${DISK_ROOT}/opt/dockerd/docker
export DISK_CACHE=${DISK_ROOT}/Cach

上面的 1 2 3 4 步骤也已经实现。

因为 adb shell 环境默认就没有下载工具。我们需要开发一个独立的软件，在 installer 文件夹中用 Golang 代码开发。
这个软件支持从 CDN 或者服务器下载 docker-*.tar.gz docker-for-android-bin-*-arm64.tar.gz 
然后解压到 /data/local/docker 里面，以及二进制解压到 /data/local/docker/bin 里面

下载之前，参考下面的代码发现硬盘时否存在：
NVME=$(mount | grep -F '/dev/block/vold/public:259,1 on /mnt/media_rw/' | grep -F ' type ext4 ' | grep -oE '/mnt/media_rw/[^ ]+')

如果硬盘不存在则报错，如果硬盘存在，则让 ${NVME}/Cache/installer 作为下载安装的临时文件夹。
设置好临时文件夹之后开始下载，并解压到上文提到的目录中。
最后 Golang 调用脚本 /data/local/docker/deploy-in-android.sh 执行，Golang 执行脚本的时候也需要把脚本的日志边执行边打印出来给用户。
执行完成则提示用户安装完成。
上文逻辑中的 installer 文件夹的代码已经实现完成。
目前需要进行优化，需要修改 installer 跟 deploy-in-android.sh 脚本。
installer 程序负责拿到硬盘路径，作为一个参数调用 deploy-in-android.sh。
而 deploy-in-android.sh 则需要接受这个参数，作为 DISK_ROOT。

请修改 deploy-in-android.sh 以及 installer 文件夹下面对应的代码。