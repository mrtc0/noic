package processes

import (
	"fmt"

	"github.com/opencontainers/runtime-spec/specs-go"
	"golang.org/x/sys/unix"
)

var rlimitMap = map[string]int{
	"RLIMIT_CPU":        unix.RLIMIT_CPU,
	"RLIMIT_FSIZE":      unix.RLIMIT_FSIZE,
	"RLIMIT_DATA":       unix.RLIMIT_DATA,
	"RLIMIT_STACK":      unix.RLIMIT_STACK,
	"RLIMIT_CORE":       unix.RLIMIT_CORE,
	"RLIMIT_RSS":        unix.RLIMIT_RSS,
	"RLIMIT_NPROC":      unix.RLIMIT_NPROC,
	"RLIMIT_NOFILE":     unix.RLIMIT_NOFILE,
	"RLIMIT_MEMLOCK":    unix.RLIMIT_MEMLOCK,
	"RLIMIT_AS":         unix.RLIMIT_AS,
	"RLIMIT_LOCKS":      unix.RLIMIT_LOCKS,
	"RLIMIT_SIGPENDING": unix.RLIMIT_SIGPENDING,
	"RLIMIT_MSGQUEUE":   unix.RLIMIT_MSGQUEUE,
	"RLIMIT_NICE":       unix.RLIMIT_NICE,
	"RLIMIT_RTPRIO":     unix.RLIMIT_RTPRIO,
	"RLIMIT_RTTIME":     unix.RLIMIT_RTTIME,
}

func strToRlimit(key string) (int, error) {
	rl, ok := rlimitMap[key]
	if !ok {
		return 0, fmt.Errorf("wrong rlimit value: %s", key)
	}
	return rl, nil
}

func SetupRlimits(pid int, process specs.Process) error {
	for _, rlimit := range process.Rlimits {
		rlimitType, err := strToRlimit(rlimit.Type)
		if err != nil {
			return err
		}

		if err := unix.Prlimit(pid, rlimitType, &unix.Rlimit{Max: rlimit.Hard, Cur: rlimit.Soft}, nil); err != nil {
			return fmt.Errorf("failed prlimit: %s", rlimit.Type)
		}
	}

	return nil
}

func SetupNowNewPrivileges() error {
	if err := unix.Prctl(unix.PR_SET_NO_NEW_PRIVS, 1, 0, 0, 0); err != nil {
		return fmt.Errorf("faild NoNewPrivileges: %s", err)
	}

	return nil
}
