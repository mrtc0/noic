package capabilities

import (
	"strings"

	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/syndtr/gocapability/capability"
)

const allCapabilityTypes = capability.CAPS | capability.BOUNDING | capability.AMBIENT

var (
	capTypes = []capability.CapType{
		capability.BOUNDING,
		capability.PERMITTED,
		capability.INHERITABLE,
		capability.EFFECTIVE,
		capability.AMBIENT,
	}

	giveCaps = map[capability.CapType][]capability.Cap{
		capability.BOUNDING:  {capability.CAP_CHOWN, capability.CAP_DAC_OVERRIDE, capability.CAP_FOWNER, capability.CAP_KILL, capability.CAP_FSETID, capability.CAP_SETGID, capability.CAP_SETUID, capability.CAP_SETFCAP},
		capability.PERMITTED: {capability.CAP_CHOWN, capability.CAP_DAC_OVERRIDE, capability.CAP_FOWNER, capability.CAP_KILL, capability.CAP_FSETID, capability.CAP_SETGID, capability.CAP_SETUID, capability.CAP_SETFCAP},
		// capability.INHERITABLE: {capability.CAP_CHOWN, capability.CAP_DAC_OVERRIDE, capability.CAP_FOWNER, capability.CAP_KILL, capability.CAP_FSETID, capability.CAP_SETGID, capability.CAP_SETUID, capability.CAP_SETFCAP},
		capability.INHERITABLE: {},
		// cap_chown,cap_dac_override,cap_fowner,cap_fsetid,cap_kill,cap_setgid,cap_setuid,cap_setpcap,cap_net_bind_service,cap_net_raw,cap_sys_chroot,cap_mknod,cap_audit_write,cap_setfcap=ep
		capability.EFFECTIVE: {capability.CAP_CHOWN, capability.CAP_DAC_OVERRIDE, capability.CAP_FOWNER, capability.CAP_KILL, capability.CAP_FSETID, capability.CAP_SETGID, capability.CAP_SETUID, capability.CAP_SETFCAP},
		capability.AMBIENT:   {capability.CAP_CHOWN, capability.CAP_DAC_OVERRIDE, capability.CAP_FOWNER, capability.CAP_KILL, capability.CAP_FSETID, capability.CAP_SETGID, capability.CAP_SETUID, capability.CAP_SETFCAP},
	}
)

type CapabilityConfig struct {
	caps map[capability.CapType][]capability.Cap
}

func newCapabilityMap() map[string]capability.Cap {
	capabilityMap := map[string]capability.Cap{}

	for _, c := range capability.List() {
		s := "CAP_" + strings.ToUpper(c.String())
		capabilityMap[s] = c
	}

	return capabilityMap
}

func stringToCapabilities(caps []string) []capability.Cap {
	capabilities := []capability.Cap{}
	capabilityMap := newCapabilityMap()

	for _, c := range caps {
		capabilities = append(capabilities, capabilityMap[c])
	}

	return capabilities
}

func New(capabilities specs.LinuxCapabilities) *CapabilityConfig {
	caps := map[capability.CapType][]capability.Cap{
		capability.BOUNDING:    stringToCapabilities(capabilities.Bounding),
		capability.AMBIENT:     stringToCapabilities(capabilities.Ambient),
		capability.EFFECTIVE:   stringToCapabilities(capabilities.Effective),
		capability.INHERITABLE: stringToCapabilities(capabilities.Inheritable),
		capability.PERMITTED:   stringToCapabilities(capabilities.Permitted),
	}

	return &CapabilityConfig{caps: caps}
}

func (c CapabilityConfig) Apply() error {

	caps, err := capability.NewPid2(0)
	if err != nil {
		return err
	}
	caps.Clear(allCapabilityTypes)

	for _, capType := range capTypes {
		caps.Set(capType, c.caps[capType]...)
	}

	if err := caps.Apply(allCapabilityTypes); err != nil {
		return err
	}

	return nil
}
