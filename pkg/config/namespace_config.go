package config

// NamespaceConfig holds the configuration for namespace selection
type NamespaceConfig struct {
	Namespaces []string
}

// NewNamespaceConfig creates a new namespace configuration
func NewNamespaceConfig(namespaces []string) *NamespaceConfig {
	if len(namespaces) == 0 {
		// Use default namespace if none specified
		namespaces = []string{"default"}
	}
	return &NamespaceConfig{
		Namespaces: namespaces,
	}
}

// GetNamespaces returns the list of namespaces to watch
func (nc *NamespaceConfig) GetNamespaces() []string {
	return nc.Namespaces
}
