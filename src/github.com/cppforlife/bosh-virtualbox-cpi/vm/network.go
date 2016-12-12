package vm

type Networks map[string]Network

type Network struct {
	Type string

	IP      string
	Netmask string
	Gateway string

	DNS     []string
	Default []string

	MAC string

	CloudPropertyName string
	CloudPropertyType string
}

func (ns Networks) Default() Network {
	var n Network

	for _, n = range ns {
		if n.IsDefaultFor("gateway") {
			break
		}
	}

	return n // returns last network
}

func (n Network) IsDefaultFor(what string) bool {
	for _, def := range n.Default {
		if def == what {
			return true
		}
	}

	return false
}

func (n Network) IsDynamic() bool { return n.Type == "dynamic" }
