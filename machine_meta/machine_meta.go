package machine_meta

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"
)

func GetMachineMeta() MachineMeta {
	m := MachineMeta{}

	hostname, _ := os.Hostname()
	m.Hostname = hostname
	m.Goos = runtime.GOOS
	m.Goarch = runtime.GOARCH

	// Get LSB (Linux Standard Base) and Distribution information.
	m.LsbRelease = LsbRelease{}
	if r, err := exec.Command("lsb_release", "-i").Output(); err == nil {
		m.LsbRelease.Id = string(r)
	}
	if r, err := exec.Command("lsb_release", "--description").Output(); err == nil {
		m.LsbRelease.Description = string(r)
	}
	if r, err := exec.Command("lsb_release", "--release").Output(); err == nil {
		m.LsbRelease.Release = string(r)
	}
	if r, err := exec.Command("lsb_release", "--codename").Output(); err == nil {
		m.LsbRelease.Codename = string(r)
	}

	m.Uname = Uname{}
	if r, err := exec.Command("uname", "--kernel-name").Output(); err == nil {
		m.Uname.KernelName = string(r)
	}
	if r, err := exec.Command("uname", "--nodename").Output(); err == nil {
		m.Uname.NodeName = string(r)
	}
	if r, err := exec.Command("uname", "--kernel-release").Output(); err == nil {
		m.Uname.KernelRelease = string(r)
	}
	if r, err := exec.Command("uname", "--kernel-version").Output(); err == nil {
		m.Uname.KernelVersion = string(r)
	}
	if r, err := exec.Command("uname", "--machine").Output(); err == nil {
		m.Uname.Machine = string(r)
	}
	if r, err := exec.Command("uname", "--processor").Output(); err == nil {
		m.Uname.Processor = string(r)
	}
	if r, err := exec.Command("uname", "--hardware-platform").Output(); err == nil {
		m.Uname.HardwarePlatform = string(r)
	}
	if r, err := exec.Command("uname", "--operating-system").Output(); err == nil {
		m.Uname.OperatingSystem = string(r)
	}
	m.AwsEc2Meta = *getAwsMeta()
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
	LsbRelease LsbRelease
	Uname      Uname
	AwsEc2Meta AwsEc2IdentityDoc
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
