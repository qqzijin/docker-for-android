#!/bin/sh

busybox mount --make-slave /
busybox mount --make-slave /sys
busybox mount --make-slave /dev
busybox mount --make-slave /proc
busybox mount --make-slave /data
busybox mount --make-slave /system 2>/dev/null

if [ -z "$DOCKER_DATA_ROOT" ]; then
	NVME=$(mount | grep -F '/dev/block/vold/public:259,1 on /mnt/media_rw/' | grep -F ' type ext4 ' | grep -oE '/mnt/media_rw/[^ ]+')
	if [ -n "$NVME" ]; then
		DOCKER_DATA_ROOT="$NVME/opt/dockerd/docker"
		mkdir -p "$DOCKER_DATA_ROOT"
		touch "$NVME/opt/.nomedia"
	fi
fi

mkdir "$DOCKER_ROOT/data" 2>/dev/null
if [ -z "$DOCKER_DATA_ROOT" ]; then
	echo "[WARN]: NVMe ext4 mountpoint not found! Use tmpfs for testing." >&2
	mount -t tmpfs -o size=100M,uid=0,gid=0,mode=0755 tmpfs "$DOCKER_ROOT/data"
else
	echo "[INFO]: NVMe docker data dir: $DOCKER_DATA_ROOT" >&2
	mount --bind "$DOCKER_DATA_ROOT" "$DOCKER_ROOT/data"
fi

unset DOCKER_DATA_ROOT

mount -t tmpfs -o size=4M,uid=0,gid=0,mode=0755 tmpfs /run

# cgroup
for controller in blkio cpuctl cpuset memcg; do
	umount /dev/$controller 2>/dev/null
done

umount /sys/fs/cgroup 2>/dev/null

TARGET_ROOTFS=

mount -t tmpfs -o size=4M,uid=0,gid=0,mode=0755 cgroup "$TARGET_ROOTFS/sys/fs/cgroup"
for controller in $(awk '!/^#/ { if ($4 == 1) print $1 }' /proc/cgroups); do
	mkdir -p "$TARGET_ROOTFS/sys/fs/cgroup/$controller"
	if ! mountpoint -q "$TARGET_ROOTFS/sys/fs/cgroup/$controller" ; then
		if ! mount -n -t cgroup -o $controller cgroup "$TARGET_ROOTFS/sys/fs/cgroup/$controller" ; then
			rmdir "$TARGET_ROOTFS/sys/fs/cgroup/$controller" || true
		fi
	fi
done

# cpuset mount with noprefix will cause docker run fail
[ -e "$TARGET_ROOTFS/sys/fs/cgroup/cpuset/cpuset.cpus" ] || {
	umount "$TARGET_ROOTFS/sys/fs/cgroup/cpuset" 2>/dev/null
	rmdir "$TARGET_ROOTFS/sys/fs/cgroup/cpuset" 2>/dev/null
}

# ssl
mkdir "$DOCKER_ROOT/var/etc-upper" 2>/dev/null
mkdir "$DOCKER_ROOT/var/etc-work" 2>/dev/null
cp -a "$DOCKER_ROOT/etc/ssl" "$DOCKER_ROOT/var/etc-upper/"
mount -o ro,noatime,lowerdir=/system/etc,upperdir=$DOCKER_ROOT/var/etc-upper,workdir=$DOCKER_ROOT/var/etc-work \
		-t overlay "overlayfs:/etc" /system/etc

if [ "$AND_DEBUG" = "1" ];then
exec sh -i
elif [ "$AND_DOCKER" = "1" ];then
dockerd --config-file "$DOCKER_ROOT/etc/docker/daemon.json" \
	--data-root "$DOCKER_ROOT/data" \
	--host "$DOCKER_HOST" \
	--pidfile "$DOCKER_ROOT/var/run/docker.pid" \
	--exec-root "$DOCKER_ROOT/var/run/docker"
else
/data/local/docker/bin/supervisord -c /data/local/docker/supervisor.conf -d
fi
