// +build darwin

package process

import (
	"bytes"
	"syscall"
	"unsafe"

	common "github.com/shirou/gopsutil/common"
	cpu "github.com/shirou/gopsutil/cpu"
	net "github.com/shirou/gopsutil/net"
)

// copied from sys/sysctl.h
const (
	CTLKern          = 1  // "high kernel": proc, limits
	KernProc         = 14 // struct: process entries
	KernProcPID      = 1  // by process id
	KernProcProc     = 8  // only return procs
	KernProcAll      = 0  // everything
	KernProcPathname = 12 // path to executable
)

type _Ctype_struct___0 struct {
	Pad uint64
}

// MemoryInfoExStat is different between OSes
type MemoryInfoExStat struct {
}

type MemoryMapsStat struct {
}

func Pids() ([]int32, error) {
	var ret []int32
	procs, err := processes()
	if err != nil {
		return ret, nil
	}

	for _, p := range procs {
		ret = append(ret, p.Pid)
	}

	return ret, nil
}

func (p *Process) Ppid() (int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}

	return k.Proc.P_pid, nil
}
func (p *Process) Name() (string, error) {
	k, err := p.getKProc()
	if err != nil {
		return "", err
	}

	return common.IntToString(k.Proc.P_comm[:]), nil
}
func (p *Process) Exe() (string, error) {
	return "", common.NotImplementedError
}
func (p *Process) Cmdline() (string, error) {
	return p.Name()
}
func (p *Process) CreateTime() (int64, error) {
	return 0, common.NotImplementedError
}
func (p *Process) Cwd() (string, error) {
	return "", common.NotImplementedError
}
func (p *Process) Parent() (*Process, error) {
	return p, common.NotImplementedError
}
func (p *Process) Status() (string, error) {
	k, err := p.getKProc()
	if err != nil {
		return "", err
	}

	return string(k.Proc.P_stat), nil // TODO
}
func (p *Process) Uids() ([]int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	uids := make([]int32, 0, 3)

	uids = append(uids, int32(k.Eproc.Pcred.P_ruid), int32(k.Eproc.Ucred.Uid), int32(k.Eproc.Pcred.P_svuid))

	return uids, nil
}
func (p *Process) Gids() ([]int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	gids := make([]int32, 0, 3)
	gids = append(gids, int32(k.Eproc.Pcred.P_rgid), int32(k.Eproc.Ucred.Ngroups), int32(k.Eproc.Pcred.P_svgid))

	return gids, nil
}
func (p *Process) Terminal() (string, error) {
	k, err := p.getKProc()
	if err != nil {
		return "", err
	}

	ttyNr := uint64(k.Eproc.Tdev)

	termmap, err := getTerminalMap()
	if err != nil {
		return "", err
	}

	return termmap[ttyNr], nil
}
func (p *Process) Nice() (int32, error) {
	k, err := p.getKProc()
	if err != nil {
		return 0, err
	}
	return int32(k.Proc.P_nice), nil
}
func (p *Process) IOnice() (int32, error) {
	return 0, common.NotImplementedError
}
func (p *Process) Rlimit() ([]RlimitStat, error) {
	var rlimit []RlimitStat
	return rlimit, common.NotImplementedError
}
func (p *Process) IOCounters() (*IOCountersStat, error) {
	return nil, common.NotImplementedError
}
func (p *Process) NumCtxSwitches() (*NumCtxSwitchesStat, error) {
	return nil, common.NotImplementedError
}
func (p *Process) NumFDs() (int32, error) {
	return 0, common.NotImplementedError
}
func (p *Process) NumThreads() (int32, error) {
	return 0, common.NotImplementedError

	/*
		k, err := p.getKProc()
		if err != nil {
			return 0, err
		}

			return k.KiNumthreads, nil
	*/
}
func (p *Process) Threads() (map[string]string, error) {
	ret := make(map[string]string, 0)
	return ret, common.NotImplementedError
}
func (p *Process) CPUTimes() (*cpu.CPUTimesStat, error) {
	return nil, common.NotImplementedError
}
func (p *Process) CPUAffinity() ([]int32, error) {
	return nil, common.NotImplementedError
}
func (p *Process) MemoryInfo() (*MemoryInfoStat, error) {
	k, err := p.getKProc()
	if err != nil {
		return nil, err
	}

	ret := &MemoryInfoStat{
		RSS: uint64(k.Eproc.Xrssize),
		VMS: uint64(k.Eproc.Xsize),
	}

	return ret, nil
}
func (p *Process) MemoryInfoEx() (*MemoryInfoExStat, error) {
	return nil, common.NotImplementedError
}
func (p *Process) MemoryPercent() (float32, error) {
	return 0, common.NotImplementedError
}

func (p *Process) Children() ([]*Process, error) {
	return nil, common.NotImplementedError
}

func (p *Process) OpenFiles() ([]OpenFilesStat, error) {
	return nil, common.NotImplementedError
}

func (p *Process) Connections() ([]net.NetConnectionStat, error) {
	return nil, common.NotImplementedError
}

func (p *Process) IsRunning() (bool, error) {
	return true, common.NotImplementedError
}
func (p *Process) MemoryMaps(grouped bool) (*[]MemoryMapsStat, error) {
	var ret []MemoryMapsStat
	return &ret, common.NotImplementedError
}

func copyParams(k *KinfoProc, p *Process) error {

	return nil
}

func processes() ([]Process, error) {
	results := make([]Process, 0, 50)

	mib := []int32{CTLKern, KernProc, KernProcAll, 0}
	buf, length, err := common.CallSyscall(mib)
	if err != nil {
		return results, err
	}

	// get kinfo_proc size
	k := KinfoProc{}
	procinfoLen := int(unsafe.Sizeof(k))
	count := int(length / uint64(procinfoLen))
	/*
		fmt.Println(length, procinfoLen, count)
		b := buf[0*procinfoLen : 0*procinfoLen+procinfoLen]
		fmt.Println(b)
		kk, err := parseKinfoProc(b)
		fmt.Printf("%#v", kk)
	*/

	// parse buf to procs
	for i := 0; i < count; i++ {
		b := buf[i*procinfoLen : i*procinfoLen+procinfoLen]
		k, err := parseKinfoProc(b)
		if err != nil {
			continue
		}
		p, err := NewProcess(int32(k.Proc.P_pid))
		if err != nil {
			continue
		}
		copyParams(&k, p)

		results = append(results, *p)
	}

	return results, nil
}

func parseKinfoProc(buf []byte) (KinfoProc, error) {
	var k KinfoProc
	br := bytes.NewReader(buf)

	err := Read(br, LittleEndian, &k)
	if err != nil {
		return k, err
	}

	return k, nil
}

func (p *Process) getKProc() (*KinfoProc, error) {
	mib := []int32{CTLKern, KernProc, KernProcPID, p.Pid}
	procK := KinfoProc{}
	length := uint64(unsafe.Sizeof(procK))
	buf := make([]byte, length)
	_, _, syserr := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&mib[0])),
		uintptr(len(mib)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&length)),
		0,
		0)
	if syserr != 0 {
		return nil, syserr
	}
	k, err := parseKinfoProc(buf)
	if err != nil {
		return nil, err
	}

	return &k, nil
}

func NewProcess(pid int32) (*Process, error) {
	p := &Process{Pid: pid}

	return p, nil
}
