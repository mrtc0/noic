package seccomp

import (
	"testing"

	"github.com/mrtc0/noic/pkg/specs"
	"github.com/stretchr/testify/assert"
)

func TestNewFilter(t *testing.T) {
	profile := specs.LinuxSeccomp{
		DefaultAction: "SCMP_ACT_ALLOW",
		Architectures: []specs.Arch{"SCMP_ARCH_X86", "SCMP_ARCH_X32"},
		Syscalls: []specs.LinuxSyscall{
			{
				Names:  []string{"getcwd", "chmod"},
				Action: "SCMP_ACT_ERRNO",
			},
		},
	}

	filter, err := NewFilter(profile)
	assert.NoError(t, err)

	assert.True(t, filter.IsValid())
}
