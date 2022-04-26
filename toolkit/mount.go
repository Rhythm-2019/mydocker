package toolkit

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const PIVOTROOT_OLD_DIR_NAME = "pivot_root"

func PivotRoot(targetRoot string) error {
	// PivotRoot 用于将某个具体目录挂载到 / 上，书主机无感知。入参为希望成为namespace中 / 的宿主机路径，
	// 需要在mount namespace下进行下面操作
	// 步骤一：产生一个 targetRoot 的挂载信息
	// 步骤二：创建一个临时目录 tmp，使用pivotRoot函数将当前 namesoace 的 / 挂载到 tmp 目录上，并将 targetRoot 挂载到 /,执行完毕后会自动将 targetRoot 变成 /（chroot）
	// 步骤三：取消 tmp 的挂载并删除该目录
	// 步骤四：将工作目录切换到 / 下

	// MS_REC 表示为子目录递归创建挂载
	if err := syscall.Mount(targetRoot, targetRoot, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("pivot root: %v", err)
	}
	pivotRoot := filepath.Join(targetRoot, PIVOTROOT_OLD_DIR_NAME)
	if err := os.MkdirAll(pivotRoot, 0777); err != nil {
		return fmt.Errorf("pivot root: mkdir pivot_root failed, detail is %v", err)
	}
	if err := syscall.PivotRoot(targetRoot, pivotRoot); err != nil {
		return fmt.Errorf("pivot root: run pivot root failed, detail is %v", err)
	}

	// MNT_DETACH：如果函数执行带有此参数，不会立即执行umount操作，而会等挂载点退出忙碌状态时才会去卸载它
	if err := syscall.Unmount(filepath.Join("/", PIVOTROOT_OLD_DIR_NAME), syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("pivot root: unmount old root failed, detail is  %v", err)
	}
	if err := os.RemoveAll(filepath.Join("/", PIVOTROOT_OLD_DIR_NAME)); err != nil {
		return fmt.Errorf("pivot root: delete old root failed, detail is  %v", err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("pivot root: chdir new root to / failed, detail is %v", err)
	}
	return nil
}

func MakeMountPoint(upperDir, workDir, lowerDir, mergeDir string) error {

	dirs := []string{upperDir, workDir, lowerDir, mergeDir}
	for _, dir := range dirs {
		err := CreateDir(dir)
		if err != nil {
			return err
		}
	}
	source := fmt.Sprintf("dirs=%s", strings.Join(dirs[:3], ":"))
	// TODO 这种方式不知道为什么报错 invalid argument
	//err := syscall.Mount(source, mergeDir, "aufs", syscall.MS_, "")
	err := exec.Command("mount", "-t", "aufs", "-o", source, "none", mergeDir).Run()
	if err != nil {
		return fmt.Errorf("mount: fail mount %s to %s in aufs, detail is %v", source, mergeDir, err)
	}

	return nil
}

func MakeBindMount(source, target string) error {
	if !HasDir(source) {
		return fmt.Errorf("mount dir %s no found", source)
	}
	if err := CreateDir(target); err != nil {
		return err
	}

	err := syscall.Mount(source, target, "bind", syscall.MS_BIND|syscall.MS_REC, "")
	if err != nil {
		return fmt.Errorf("mount: fail mount %s to %s in bind, detail is %v", source, target, err)
	}
	return nil
}
func Umount(target string) error {
	return syscall.Unmount(target, 0)
}

func MountProc() error {
	if err := CreateDir("/proc"); err != nil {
		return err
	}
	return syscall.Mount("proc", "/proc", "proc", uintptr(syscall.MS_NOEXEC|syscall.MS_NODEV|syscall.MS_NOSUID), "")
}
func MountDev() error {
	if err := CreateDir("/dev"); err != nil {
		return err
	}
	return syscall.Mount("tmpfs", "/dev", "tmpfs", uintptr(syscall.MS_NOEXEC|syscall.MS_STRICTATIME), "")
}
