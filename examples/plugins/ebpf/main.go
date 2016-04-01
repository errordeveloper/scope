package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func main() {
	hostname, _ := os.Hostname()
	var (
		addr   = flag.String("addr", "/var/run/scope/plugins/ebpf.sock", "unix socket to listen for connections on")
		hostID = flag.String("hostname", hostname, "hostname of the host running this plugin")
	)
	flag.Parse()

	log.Println("Starting...")

	// Check we have ebpf available

	// Compile and load the ebpf script

	os.Remove(*addr)
	listener, err := net.Listen("unix", *addr)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		listener.Close()
		os.Remove(*addr)
	}()

	log.Printf("Listening on: unix://%s", *addr)

	plugin := &Plugin{HostID: *hostID}
	http.HandleFunc("/", plugin.Handshake)
	http.HandleFunc("/report", plugin.Report)
	if err := http.Serve(listener, nil); err != nil {
		log.Printf("error: %v", err)
	}
}

type Plugin struct {
	HostID string
}

func (p *Plugin) Handshake(w http.ResponseWriter, r *http.Request) {
	log.Printf("Probe %s handshake", r.FormValue("probe_id"))
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"name":        "ebpf",
		"description": "Adds a graph of http requests/second to processes",
		"interfaces":  []string{"reporter"},
		"api_version": "1",
	})
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func (p *Plugin) Report(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	nowISO := now.Format(time.RFC3339)
	value, err := iowait()
	if err != nil {
		log.Printf("error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"Host": map[string]interface{}{
			"nodes": map[string]interface{}{
				p.HostID + ";<host>": map[string]interface{}{
					"metrics": map[string]interface{}{
						"iowait": map[string]interface{}{
							"samples": []interface{}{
								map[string]interface{}{
									"date":  nowISO,
									"value": value,
								},
							},
						},
					},
				},
			},
			"metric_templates": map[string]interface{}{
				"iowait": map[string]interface{}{
					"id":       "iowait",
					"label":    "IO Wait",
					"format":   "percent",
					"priority": 0.1, // low number so it shows up first
				},
			},
		},
	})
	if err != nil {
		log.Printf("error: %v", err)
	}
}

// Get the latest iowait value
func iowait() (float64, error) {
	out, err := exec.Command("iostat", "-c").Output()
	if err != nil {
		return 0, fmt.Errorf("iowait: %v", err)
	}

	// Linux 4.2.0-25-generic (a109563eab38)	04/01/16	_x86_64_(4 CPU)
	//
	// avg-cpu:  %user   %nice %system %iowait  %steal   %idle
	//	          2.37    0.00    1.58    0.01    0.00   96.04
	lines := strings.Split(string(out), "\n")
	if len(lines) < 4 {
		return 0, fmt.Errorf("iowait: unexpected output: %q", out)
	}

	values := strings.Fields(lines[3])
	if len(values) != 6 {
		return 0, fmt.Errorf("iowait: unexpected output: %q", out)
	}

	return strconv.ParseFloat(values[3], 64)
}
