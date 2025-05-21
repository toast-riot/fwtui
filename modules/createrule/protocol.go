package createrule

type Protocol string

const (
	ProtocolTcp  Protocol = "tcp"
	ProtocolUdp  Protocol = "udp"
	ProtocolBoth Protocol = "tcp/udp"
)

var protocols = []Protocol{ProtocolBoth, ProtocolTcp, ProtocolUdp}
