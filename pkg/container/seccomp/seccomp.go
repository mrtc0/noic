package seccomp

import (
	"fmt"
	"syscall"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	libseccomp "github.com/seccomp/libseccomp-golang"
)

const defaultErrnoRetCode = uint(syscall.EPERM)

var archs = map[string]string{
	"SCMP_ARCH_X86":         "x86",
	"SCMP_ARCH_X86_64":      "amd64",
	"SCMP_ARCH_X32":         "x32",
	"SCMP_ARCH_ARM":         "arm",
	"SCMP_ARCH_AARCH64":     "arm64",
	"SCMP_ARCH_MIPS":        "mips",
	"SCMP_ARCH_MIPS64":      "mips64",
	"SCMP_ARCH_MIPS64N32":   "mips64n32",
	"SCMP_ARCH_MIPSEL":      "mipsel",
	"SCMP_ARCH_MIPSEL64":    "mipsel64",
	"SCMP_ARCH_MIPSEL64N32": "mipsel64n32",
	"SCMP_ARCH_PPC":         "ppc",
	"SCMP_ARCH_PPC64":       "ppc64",
	"SCMP_ARCH_PPC64LE":     "ppc64le",
	"SCMP_ARCH_S390":        "s390",
	"SCMP_ARCH_S390X":       "s390x",
}

func LoadSeccompProfile(profile specs.LinuxSeccomp) error {
	filter, err := NewFilter(profile)
	if err != nil {
		return err
	}

	if err := filter.SetNoNewPrivsBit(false); err != nil {
		return err
	}

	return filter.Load()
}

func NewFilter(profile specs.LinuxSeccomp) (*libseccomp.ScmpFilter, error) {
	errnoRet := defaultErrnoRet(profile)

	action, err := FindLibSeccompScmpAction(string(profile.DefaultAction), &errnoRet)
	if err != nil {
		return nil, fmt.Errorf("invalid seccomp action %s: %s", action, err)
	}

	filter, err := libseccomp.NewFilter(action)
	if err != nil {
		return nil, fmt.Errorf("failed creating seccomp filter: %s", err)
	}

	for _, arch := range profile.Architectures {
		scmpArch, err := libseccomp.GetArchFromString(archs[string(arch)])
		if err != nil {
			return nil, fmt.Errorf("invalid architecture %s: %s", string(arch), err)
		}

		if err := filter.AddArch(scmpArch); err != nil {
			return nil, fmt.Errorf("failed add architecture to filter: %s", err)
		}
	}

	for _, s := range profile.Syscalls {
		if len(s.Names) == 0 {
			return nil, fmt.Errorf("syscalls is empty")
		}
		for _, name := range s.Names {
			syscallID, err := libseccomp.GetSyscallFromName(name)
			if err != nil {
				return nil, fmt.Errorf("invalid syscall. %s is not found: %s", name, err)
			}

			if s.ErrnoRet != nil {
				errnoRet = *s.ErrnoRet
			}

			action, err := FindLibSeccompScmpAction(string(s.Action), &errnoRet)
			if err != nil {
				return nil, fmt.Errorf("invalid seccomp action %s: %s", action, err)
			}

			filter.AddRule(syscallID, action)
		}
	}

	return filter, err
}

func defaultErrnoRet(profile specs.LinuxSeccomp) uint {
	if profile.DefaultErrnoRet != nil {
		return *profile.DefaultErrnoRet
	}

	return defaultErrnoRetCode
}

func FindLibSeccompScmpAction(action string, errno *uint) (libseccomp.ScmpAction, error) {
	switch action {
	case "SCMP_ACT_KILL":
		return libseccomp.ActKill, nil
	case "SCMP_ACT_KILL_PROCESS":
		return libseccomp.ActKillProcess, nil
	case "SCMP_ACT_KILL_THREAD":
		return libseccomp.ActKillThread, nil
	case "SCMP_ACT_TRAP":
		return libseccomp.ActTrap, nil
	case "SCMP_ACT_ERRNO":
		return libseccomp.ActErrno.SetReturnCode(int16(*errno)), nil
	case "SCMP_ACT_TRACE":
		return libseccomp.ActTrace.SetReturnCode(int16(*errno)), nil
	case "SCMP_ACT_ALLOW":
		return libseccomp.ActAllow, nil
	case "SCMP_ACT_LOG":
		return libseccomp.ActLog, nil
	case "SCMP_ACT_NOTIFY":
		return libseccomp.ActNotify, nil
	}

	return libseccomp.ActInvalid, fmt.Errorf("invalid seccomp action")

}
