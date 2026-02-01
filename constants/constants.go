package constants

const (
	// Protocols and Ports
	RelayPort   = "9091"
	WSPort      = "9092"
	WebRTCPort  = "9093"
	HTTPAPIPort = "9094"

	// PubSub Topics (Must match JS constants)
	DiscoveryTopic   = "_peer-discovery._p2p._pubsub"
	SpreadsheetTopic = "spreadsheet-1"

	DefaultKeyPath = "keys/relay-peer.key"
)
