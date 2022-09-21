package specs

type Process struct {
	Args []string
}

func SetupSpec() *Process {
	spec := &Process{}
	return spec
}
