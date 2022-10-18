package seccomp

import (
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/stretchr/testify/assert"
)

func TestNewFilter_ValidCase(t *testing.T) {
	testCases := []struct {
		name    string
		profile specs.LinuxSeccomp
	}{
		{
			name: "sample profile",
			profile: specs.LinuxSeccomp{
				DefaultAction: "SCMP_ACT_ALLOW",
				Architectures: []specs.Arch{"SCMP_ARCH_X86", "SCMP_ARCH_X32"},
				Syscalls: []specs.LinuxSyscall{
					{
						Names:  []string{"getcwd", "chmod"},
						Action: "SCMP_ACT_ERRNO",
					},
				},
			},
		},
		{
			name: "arch is empty",
			profile: specs.LinuxSeccomp{
				DefaultAction: "SCMP_ACT_ALLOW",
				Syscalls: []specs.LinuxSyscall{
					{
						Names:  []string{"getcwd", "chmod"},
						Action: "SCMP_ACT_ERRNO",
					},
				},
			},
		},
		{
			name: "syscall is empty",
			profile: specs.LinuxSeccomp{
				DefaultAction: "SCMP_ACT_ALLOW",
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			filter, err := NewFilter(test.profile)
			assert.NoError(t, err)

			assert.True(t, filter.IsValid())
		})
	}
}

func TestNewFilter_InvalidCase(t *testing.T) {
	testCases := []struct {
		name    string
		profile specs.LinuxSeccomp
	}{
		{
			name: "default action is empty",
			profile: specs.LinuxSeccomp{
				Architectures: []specs.Arch{"SCMP_ARCH_X86", "SCMP_ARCH_X32"},
				Syscalls: []specs.LinuxSyscall{
					{
						Names:  []string{"getcwd", "chmod"},
						Action: "SCMP_ACT_ERRNO",
					},
				},
			},
		},
		{
			name: "syscall names is empty",
			profile: specs.LinuxSeccomp{
				DefaultAction: "SCMP_ACT_ALLOW",
				Syscalls: []specs.LinuxSyscall{
					{
						Action: "SCMP_ACT_ERRNO",
					},
				},
			},
		},
		{
			name: "syscall action is empty",
			profile: specs.LinuxSeccomp{
				DefaultAction: "SCMP_ACT_ALLOW",
				Syscalls: []specs.LinuxSyscall{
					{
						Names: []string{"getcwd", "chmod"},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewFilter(test.profile)
			assert.Error(t, err)
		})
	}
}
