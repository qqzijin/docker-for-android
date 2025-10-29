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
版本号我们跟随 docker 版本以及加上我们的子版本号。目前的版本确定为：28.0.1.10
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

为了方便开发，我们需要实现一个脚本，从默认的服务器拿到版本号，以及版本号里面的信息，并下载 arm64_bin 的二进制文件到 arm64_bin 当中。

请实现从CDN中判断 version.txt ，并根据文件信息下载 arm64_bin 并把文件覆盖到 arm64_bin 目录中

