package drives

import "testing"

const partitionMountInfo = `22 1 259:2 / / rw - ext4 /dev/root rw
23 22 7:0 / /snap/core/1 ro - squashfs /dev/loop0 ro
24 22 259:1 / /boot/efi rw - vfat /dev/nvme0n1p1 rw
36 22 259:3 / /mnt/6AB2DBF3B2DBC1AD rw - ntfs3 /dev/nvme0n1p3 rw
37 22 259:3 /Games /mnt/Games rw - ntfs3 /dev/nvme0n1p3 rw
40 22 259:4 / /mnt/My\040Games rw - ext4 /dev/sdb1 rw
41 22 0:45 / /run/user/1000/gvfs rw - fuse.gvfsd-fuse gvfsd-fuse rw
`

func TestPartitionsFromMountInfoKeepsUserDisksOnce(t *testing.T) {
	partitions := partitionsFromMountInfo(partitionMountInfo)

	if len(partitions) != 2 {
		t.Fatalf("got %d partitions, want 2: %+v", len(partitions), partitions)
	}
	first := partitions[0]
	if first.Device != "/dev/nvme0n1p3" || first.Mount != "/mnt/6AB2DBF3B2DBC1AD" || first.FsType != "ntfs3" {
		t.Errorf("first partition = %+v", first)
	}
	if second := partitions[1]; second.Mount != "/mnt/My Games" {
		t.Errorf("second mount = %q, want the unescaped name", second.Mount)
	}
}

func TestPartitionsFromMountInfoIsEmptyWithoutMounts(t *testing.T) {
	if got := partitionsFromMountInfo(""); got == nil || len(got) != 0 {
		t.Errorf("got %#v, want an empty slice", got)
	}
}
