package panel

import (
	"bufio"
	"fmt"
	"github.com/Yuzuki616/V2bX/conf"
	"github.com/go-resty/resty/v2"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Panel is the interface for different panel's api.

type ClientInfo struct {
	APIHost  string
	NodeID   int
	Key      string
	NodeType string
}

type Client struct {
	client        *resty.Client
	APIHost       string
	Key           string
	NodeType      string
	NodeId        int
	SpeedLimit    int
	DeviceLimit   int
	LocalRuleList []DestinationRule
	etag          string
}

func New(c *conf.ApiConfig) (Panel, error) {
	client := resty.New()
	client.SetRetryCount(3)
	if c.Timeout > 0 {
		client.SetTimeout(time.Duration(c.Timeout) * time.Second)
	} else {
		client.SetTimeout(5 * time.Second)
	}
	client.OnError(func(req *resty.Request, err error) {
		if v, ok := err.(*resty.ResponseError); ok {
			// v.Response contains the last response from the server
			// v.Err contains the original error
			log.Print(v.Err)
		}
	})
	client.SetBaseURL(c.APIHost)
	// Check node type
	if c.NodeType != "V2ray" &&
		c.NodeType != "Trojan" &&
		c.NodeType != "Shadowsocks" {
		return nil, fmt.Errorf("unsupported Node type: %s", c.NodeType)
	}
	// Create Key for each requests
	client.SetQueryParams(map[string]string{
		"node_type": strings.ToLower(c.NodeType),
		"node_id":   strconv.Itoa(c.NodeID),
		"token":     c.Key,
	})
	// Read local rule list
	localRuleList := readLocalRuleList(c.RuleListPath)
	return &Client{
		client:        client,
		Key:           c.Key,
		APIHost:       c.APIHost,
		NodeType:      c.NodeType,
		SpeedLimit:    c.SpeedLimit,
		DeviceLimit:   c.DeviceLimit,
		NodeId:        c.NodeID,
		LocalRuleList: localRuleList,
	}, nil
}

// readLocalRuleList reads the local rule list file
func readLocalRuleList(path string) (LocalRuleList []DestinationRule) {
	LocalRuleList = make([]DestinationRule, 0)
	if path != "" {
		// open the file
		file, err := os.Open(path)
		//handle errors while opening
		if err != nil {
			log.Printf("Error when opening file: %s", err)
			return
		}
		fileScanner := bufio.NewScanner(file)
		// read line by line
		for fileScanner.Scan() {
			LocalRuleList = append(LocalRuleList, DestinationRule{
				ID:      -1,
				Pattern: regexp.MustCompile(fileScanner.Text()),
			})
		}
		// handle first encountered error while reading
		if err := fileScanner.Err(); err != nil {
			log.Fatalf("Error while reading file: %s", err)
			return
		}
	}
	return
}
