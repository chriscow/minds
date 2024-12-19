package minds

type Metadata map[string]any

func (m Metadata) Copy() Metadata {
	copied := make(Metadata, len(m))
	for k, v := range m {
		copied[k] = v
	}
	return copied
}
