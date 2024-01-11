package dockerutils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadContainerID(t *testing.T) {
	for in, out := range map[string]string{
		`
955 729 0:119 / / rw,relatime master:565 - overlay overlay rw,lowerdir=/var/lib/docker/overlay2/l/WQPTXVGXG6ZVSXNO2VW27URZCO:/var/lib/docker/overlay2/l/INM3CVMNWSJ2WHTRFKPCBEKIY7,upperdir=/var/lib/docker/overlay2/ffe2928d424a97b0316886864a6d97ed8201620d9e834c7bdf8a678272eacbce/diff,workdir=/var/lib/docker/overlay2/ffe2928d424a97b0316886864a6d97ed8201620d9e834c7bdf8a678272eacbce/work,index=off
956 955 0:122 / /proc rw,nosuid,nodev,noexec,relatime - proc proc rw
957 955 0:123 / /dev rw,nosuid - tmpfs tmpfs rw,size=65536k,mode=755,inode64
958 957 0:124 / /dev/pts rw,nosuid,noexec,relatime - devpts devpts rw,gid=5,mode=620,ptmxmode=666
959 955 0:125 / /sys ro,nosuid,nodev,noexec,relatime - sysfs sysfs ro
960 959 0:27 / /sys/fs/cgroup ro,nosuid,nodev,noexec,relatime - cgroup2 cgroup rw
961 957 0:121 / /dev/mqueue rw,nosuid,nodev,noexec,relatime - mqueue mqueue rw
962 957 0:126 / /dev/shm rw,nosuid,nodev,noexec,relatime - tmpfs shm rw,size=65536k,inode64
963 955 8:4 /var/lib/docker/containers/7accf056f96dcfdeb486c705843147f979d3e0ce88f7375e145688e0d2890e33/resolv.conf /etc/resolv.conf rw,relatime - ext4 /dev/sda4 rw
964 955 8:4 /var/lib/docker/containers/7accf056f96dcfdeb486c705843147f979d3e0ce88f7375e145688e0d2890e33/hostname /etc/hostname rw,relatime - ext4 /dev/sda4 rw
965 955 8:4 /var/lib/docker/containers/7accf056f96dcfdeb486c705843147f979d3e0ce88f7375e145688e0d2890e33/hosts /etc/hosts rw,relatime - ext4 /dev/sda4 rw
730 957 0:124 /0 /dev/console rw,nosuid,noexec,relatime - devpts devpts rw,gid=5,mode=620,ptmxmode=666
731 956 0:122 /bus /proc/bus ro,nosuid,nodev,noexec,relatime - proc proc rw
732 956 0:122 /fs /proc/fs ro,nosuid,nodev,noexec,relatime - proc proc rw
910 956 0:122 /irq /proc/irq ro,nosuid,nodev,noexec,relatime - proc proc rw
911 956 0:122 /sys /proc/sys ro,nosuid,nodev,noexec,relatime - proc proc rw
912 956 0:122 /sysrq-trigger /proc/sysrq-trigger ro,nosuid,nodev,noexec,relatime - proc proc rw
913 956 0:127 / /proc/asound ro,relatime - tmpfs tmpfs ro,inode64
914 956 0:128 / /proc/acpi ro,relatime - tmpfs tmpfs ro,inode64
915 956 0:123 /null /proc/kcore rw,nosuid - tmpfs tmpfs rw,size=65536k,mode=755,inode64
916 956 0:123 /null /proc/keys rw,nosuid - tmpfs tmpfs rw,size=65536k,mode=755,inode64
917 956 0:123 /null /proc/latency_stats rw,nosuid - tmpfs tmpfs rw,size=65536k,mode=755,inode64
918 956 0:123 /null /proc/timer_list rw,nosuid - tmpfs tmpfs rw,size=65536k,mode=755,inode64
919 956 0:129 / /proc/scsi ro,relatime - tmpfs tmpfs ro,inode64
920 959 0:130 / /sys/firmware ro,relatime - tmpfs tmpfs ro,inode64
921 959 0:131 / /sys/devices/virtual/powercap ro,relatime - tmpfs tmpfs ro,inode64`: "7accf056f96dcfdeb486c705843147f979d3e0ce88f7375e145688e0d2890e33",
	} {
		id, err := parseContainerIDv2(strings.NewReader(in))
		assert.NoError(t, err)
		assert.Equal(t, out, id)
	}
}
