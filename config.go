package raftify

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/memberlist"

	"github.com/hashicorp/logutils"
)

// Timeout and ticker settings for maximum performance.
const (
	// Time interval measured in milliseconds in which candidates send out
	// vote requests and leaders send out heartbeats.
	TickerInterval = 200

	// Minimum time measured in milliseconds that a non-leader waits for a
	// current heartbeat to arrive before switching to a precandidate.
	MinTimeout = 800

	// Maximum time measured in milliseconds that a non-leader waits for a
	// current heartbeat to arrive before switching to a precandidate.
	MaxTimeout = 1200

	// Maximum number of cycles a leader can go without getting heartbeat
	// responses from a majority of cluster members. If there are not enough
	// heartbeat responses for more than cycles than this value, the leader
	// is forced to step down.
	MaxSubQuorumCycles = int(MinTimeout/TickerInterval) - 1

	// Maximum number of cycles a precandidate couldn't reach the quorum
	// in order to become a candidate. If this threshold is met or exceeded
	// it is safe to assume the cluster suffers a network partition and the
	// node in question is partitioned out into a smaller cluster that can
	// never reach the quorum. Upon reaching this threshold, a rejoin event
	// is triggered to make the node in question aware of the network partition.
	MaxMissedPrevoteCycles = 5
)

// Config contains the contents of the raftify.json file.
type Config struct {
	// Mandatory. The unique identifier of a node.
	ID string `json:"id"`

	// Mandatory. Self-imposed limit of nodes that can be run in one cluster.
	// This is needed to allocate enough memory for the buffered channel used
	// for event messages.
	MaxNodes int `json:"max_nodes"`

	// Mandatory. The 16-, 24- or 32-byte AES encryption key used to encrypt
	// the message exchange between cluster members.
	Encrypt string `json:"encrypt"`

	// The performance multiplier that determines how the timeouts and
	// intervals scale. This can be used to adjust the timeout settings
	// for higher latency environments.
	Performance int `json:"performance"`

	// The number of expected nodes to go online before starting the
	// Raft leader election and bootstrapping the cluster.
	Expect int `json:"expect"`

	// The log levels for raftify; can be DEBUG, INFO, WARN or ERR.
	LogLevel string `json:"log_level"`

	// The address to bind the node to.
	BindAddr string `json:"bind_addr"`

	// The port to bind the node to.
	BindPort int `json:"bind_port"`

	// The list of peers to contact in order to join an existing cluster
	// or form a new one.
	PeerList []string `json:"peer_list"`
}

// truncPeerList removes the local node from the peerlist.
func (c *Config) truncPeerList(address string) error {
	for i := range c.PeerList {
		if c.PeerList[i] == address {
			c.PeerList = append(c.PeerList[:i], c.PeerList[i+1:]...)
			return nil
		}
	}
	return errors.New("local node is not listed as a peer")
}

// validate checks for constraint violations in the raftify.json file.
func (c *Config) validate() error {
	// Variable used to aggregate all errors found during validation.
	var errs string

	// Set defaults.
	if c.Performance == 0 {
		c.Performance = 1
	}
	if c.LogLevel == "" {
		c.LogLevel = "WARN"
	}
	if c.BindAddr == "" {
		c.BindAddr = "0.0.0.0"
	}
	if c.BindPort == 0 {
		c.BindPort = 7946
	}

	// Check constraints.
	if c.ID == "" {
		errs += "\tid must not be empty\n"
	}
	if c.MaxNodes <= 0 {
		errs += "\tmax_nodes must be greater than 0\n"
	}
	if secretKey, err := hexToByte(c.Encrypt); err != nil {
		if err := memberlist.ValidateKey(secretKey); err != nil {
			errs += fmt.Sprintf("\tencrypt must be of length 16, 24 or 32 bytes: got %v bytes\n", len(secretKey))
		}
	}
	if c.Performance < 0 {
		errs += "\tperformance must be greater than 0\n"
	}
	if c.Expect < 1 || c.Expect > c.MaxNodes {
		errs += fmt.Sprintf("\texpect must be between 1 and %v\n", c.MaxNodes)
	}
	if c.Expect > 1 && len(c.PeerList) == 0 {
		errs += "\tpeerlist must not be empty if more than one node is expected for bootstrap\n"
	}
	if match, _ := regexp.MatchString(`DEBUG|INFO|WARN|ERR`, c.LogLevel); !match {
		errs += "\tlog_level must be DEBUG, INFO, WARN or ERR\n"
	}
	if ip := net.ParseIP(c.BindAddr); ip == nil {
		errs += "\tbind_addr is not a valid IPv4\n"
	}
	if c.BindPort < 0 || c.BindPort > 65535 {
		errs += fmt.Sprintf("\tbind_port %v must be in range 0-65535\n", c.BindPort)
	}
	if len(c.PeerList) > c.MaxNodes {
		errs += fmt.Sprintf("\tpeer_list must not contain more than %v peers, including the local node: got %v peers\n", c.MaxNodes, len(c.PeerList))
	}
	for _, peer := range c.PeerList {
		host, port, err := net.SplitHostPort(peer)
		if err != nil {
			errs += fmt.Sprintf("\tpeer address %v is not a valid host:port address\n", peer)
		}
		if ip := net.ParseIP(host); ip == nil {
			errs += fmt.Sprintf("\tbind_addr %v is not a valid IPv4\n", host)
		}
		if p, _ := strconv.Atoi(port); p < 0 || p > 65535 {
			errs += fmt.Sprintf("\tbind_port %v must be in range 0-65535\n", port)
		}
	}

	if errs != "" {
		return fmt.Errorf("found errors in raftify.json:\n%v", strings.TrimSuffix(errs, "\n"))
	}
	return nil
}

// loadConfig loads the contents of the raftify.json file into memory.
func (n *Node) loadConfig() error {
	configJSON, err := os.Open(n.workingDir + "/raftify.json")
	if err != nil {
		return err
	}
	defer configJSON.Close()

	configBytes, err := ioutil.ReadAll(configJSON)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(configBytes, &n.config); err != nil {
		return err
	}

	// If the node needs to rejoin, overwrite the peerlist from the raftify.json
	// with the memberlist persisted in the state.json file.
	if n.rejoin {
		n.logger.Println("[DEBUG] raftify: Preparing to rejoin the cluster...")

		list, err := n.loadState()
		if err != nil {
			return err
		}

		n.config.PeerList = []string{}
		localNode := fmt.Sprintf("%v:%v", n.config.BindAddr, n.config.BindPort)
		for _, node := range list {
			if node.Address() == localNode {
				continue
			}
			n.config.PeerList = append(n.config.PeerList, node.Address())
		}
	}

	// Remove local node from peerlist such that the join event throws an error if none of
	// the other peers could be reached. It needs to be removed because it will always reach
	// itself which is obvious on one hand and not necessary on the other.
	// Also, the truncation needs to be done before the validation.
	n.config.truncPeerList(fmt.Sprintf("%v:%v", n.config.BindAddr, n.config.BindPort))

	if err = n.config.validate(); err != nil {
		return err
	}

	n.logger.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERR"},
		MinLevel: logutils.LogLevel(n.config.LogLevel),
		Writer:   os.Stderr,
	})
	return nil
}
