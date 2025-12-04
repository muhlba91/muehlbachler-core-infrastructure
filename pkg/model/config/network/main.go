package network

// Config defines network configuration.
type Config struct {
	// Name is the name of the network.
	Name *string `yaml:"name,omitempty"`
	// DNSSuffix is the DNS suffix for the network.
	DNSSuffix *string `yaml:"dnsSuffix,omitempty"`
	// CIDR is the CIDR block for the network.
	CIDR *string `yaml:"cidr,omitempty"`
	// SubnetCIDR is the CIDR block for the subnet.
	SubnetCIDR *string `yaml:"subnetCidr,omitempty"`
	// FirewallRules are the firewall rules for the network.
	FirewallRules map[string]*FirewallRule `yaml:"firewallRules,omitempty"`
}

// FirewallRule defines a firewall rule.
type FirewallRule struct {
	// Description is the description of the firewall rule.
	Description *string `yaml:"description,omitempty"`
	// Port is the port of the firewall rule.
	Port *int `yaml:"port,omitempty"`
	// Protocol is the protocol of the firewall rule.
	Protocol *string `yaml:"protocol,omitempty"` // 'tcp' | 'udp' | 'icmp'
	// SourceIPs are the source IPs of the firewall rule.
	SourceIPs []string `yaml:"sourceIPs,omitempty"`
}
