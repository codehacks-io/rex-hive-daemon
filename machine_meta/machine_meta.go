package machine_meta

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

func GetMachineMeta() *MachineMeta {
	m := &MachineMeta{}

	hostname, _ := os.Hostname()
	m.Hostname = hostname
	m.Goos = runtime.GOOS
	m.Goarch = runtime.GOARCH

	// Get LSB (Linux Standard Base) and Distribution information.
	if m.Goos == "linux" {
		m.LsbRelease = &LsbRelease{}
		if r, err := exec.Command("lsb_release", "--id").Output(); err == nil {
			m.LsbRelease.Id = strings.Trim(string(r), "\n")
		}
		if r, err := exec.Command("lsb_release", "--description").Output(); err == nil {
			m.LsbRelease.Description = strings.Trim(string(r), "\n")
		}
		if r, err := exec.Command("lsb_release", "--release").Output(); err == nil {
			m.LsbRelease.Release = strings.Trim(string(r), "\n")
		}
		if r, err := exec.Command("lsb_release", "--codename").Output(); err == nil {
			m.LsbRelease.Codename = strings.Trim(string(r), "\n")
		}
		m.Uname = &Uname{}
		if r, err := exec.Command("uname", "--kernel-name").Output(); err == nil {
			m.Uname.KernelName = strings.Trim(string(r), "\n")
		}
		if r, err := exec.Command("uname", "--nodename").Output(); err == nil {
			m.Uname.NodeName = strings.Trim(string(r), "\n")
		}
		if r, err := exec.Command("uname", "--kernel-release").Output(); err == nil {
			m.Uname.KernelRelease = strings.Trim(string(r), "\n")
		}
		if r, err := exec.Command("uname", "--kernel-version").Output(); err == nil {
			m.Uname.KernelVersion = strings.Trim(string(r), "\n")
		}
		if r, err := exec.Command("uname", "--machine").Output(); err == nil {
			m.Uname.Machine = strings.Trim(string(r), "\n")
		}
		if r, err := exec.Command("uname", "--processor").Output(); err == nil {
			m.Uname.Processor = strings.Trim(string(r), "\n")
		}
		if r, err := exec.Command("uname", "--hardware-platform").Output(); err == nil {
			m.Uname.HardwarePlatform = strings.Trim(string(r), "\n")
		}
		if r, err := exec.Command("uname", "--operating-system").Output(); err == nil {
			m.Uname.OperatingSystem = strings.Trim(string(r), "\n")
		}
	}
	m.AwsEc2Meta = getAwsMeta()
	return m
}

func getAwsMeta() *AwsEc2IdentityDoc {
	res, err := http.Get("http://169.254.169.254/latest/dynamic/instance-identity/document") // Constant URL
	if err != nil {
		fmt.Println("Cannot call AWS info endpoint. Program might not be running on AWS")
		return nil
	}

	if res.StatusCode != 200 {
		fmt.Println(fmt.Sprintf("Called call AWS endpoint but received status code %d", res.StatusCode))
		return nil
	}

	b, _ := io.ReadAll(res.Body)

	doc := &AwsEc2IdentityDoc{}
	_ = json.Unmarshal(b, doc)
	return doc
}

type LsbRelease struct {
	Id          string // eg: Distributor ID: Ubuntu
	Description string // eg: Description:    Ubuntu 20.04.5 LTS
	Release     string // eg: Release:        20.04
	Codename    string // eg: Codename:       focal
}
type Uname struct {
	KernelName       string // eg: Linux
	NodeName         string // eg: GODLIKE
	KernelRelease    string // eg: 5.10.16.3-microsoft-standard-WSL2
	KernelVersion    string // eg: #1 SMP Fri Apr 2 22:23:49 UTC 2021
	Machine          string // eg: x86_64
	Processor        string // eg: x86_64
	HardwarePlatform string // eg: x86_64
	OperatingSystem  string // eg: GNU/Linux
}

type MachineMeta struct {
	Hostname   string
	Goos       string
	Goarch     string
	LsbRelease *LsbRelease
	Uname      *Uname
	AwsEc2Meta *AwsEc2IdentityDoc
}

type AwsEc2IdentityDoc struct {
	AccountId               string      `json:"accountId"`
	Architecture            string      `json:"architecture"`
	AvailabilityZone        string      `json:"availabilityZone"`
	BillingProducts         interface{} `json:"billingProducts"`
	DevPayProductCodes      interface{} `json:"devpayProductCodes"`
	MarketplaceProductCodes interface{} `json:"marketplaceProductCodes"`
	ImageId                 string      `json:"imageId"`
	InstanceId              string      `json:"instanceId"`
	InstanceType            string      `json:"instanceType"`
	KernelId                interface{} `json:"kernelId"`
	PendingTime             time.Time   `json:"pendingTime"`
	PrivateIp               string      `json:"privateIp"`
	RamdiskId               interface{} `json:"ramdiskId"`
	Region                  string      `json:"region"`
	Version                 string      `json:"version"`
}
