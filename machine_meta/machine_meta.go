package machine_meta

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
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
		re := regexp.MustCompile(`(.+:\s*)`)
		m.LsbRelease = &LsbRelease{}
		if r, err := exec.Command("lsb_release", "--id").Output(); err == nil {
			m.LsbRelease.Id = re.ReplaceAllString(strings.Trim(string(r), "\n"), "")
		}
		if r, err := exec.Command("lsb_release", "--description").Output(); err == nil {
			m.LsbRelease.Description = re.ReplaceAllString(strings.Trim(string(r), "\n"), "")
		}
		if r, err := exec.Command("lsb_release", "--release").Output(); err == nil {
			m.LsbRelease.Release = re.ReplaceAllString(strings.Trim(string(r), "\n"), "")
		}
		if r, err := exec.Command("lsb_release", "--codename").Output(); err == nil {
			m.LsbRelease.Codename = re.ReplaceAllString(strings.Trim(string(r), "\n"), "")
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
	c := http.Client{}
	c.Timeout = time.Second * 2
	res, err := c.Get("http://169.254.169.254/latest/dynamic/instance-identity/document") // Constant URL
	if err != nil {
		fmt.Println("Cannot call AWS EC2 info endpoint, might not be running on AWS.")
		return nil
	}

	if res.StatusCode != 200 {
		fmt.Println(fmt.Sprintf("Called call AWS EC2 info endpoint but received status code %d", res.StatusCode))
		return nil
	}

	b, _ := io.ReadAll(res.Body)

	doc := &AwsEc2IdentityDoc{}
	_ = json.Unmarshal(b, doc)
	return doc
}

type LsbRelease struct {
	Id          string `bson:"id,omitempty"`          // eg: Distributor ID: Ubuntu
	Description string `bson:"description,omitempty"` // eg: Description:    Ubuntu 20.04.5 LTS
	Release     string `bson:"release,omitempty"`     // eg: Release:        20.04
	Codename    string `bson:"codename,omitempty"`    // eg: Codename:       focal
}
type Uname struct {
	KernelName       string `bson:"kernelName,omitempty"`       // eg: Linux
	NodeName         string `bson:"nodeName,omitempty"`         // eg: GODLIKE
	KernelRelease    string `bson:"kernelRelease,omitempty"`    // eg: 5.10.16.3-microsoft-standard-WSL2
	KernelVersion    string `bson:"kernelVersion,omitempty"`    // eg: #1 SMP Fri Apr 2 22:23:49 UTC 2021
	Machine          string `bson:"machine,omitempty"`          // eg: x86_64
	Processor        string `bson:"processor,omitempty"`        // eg: x86_64
	HardwarePlatform string `bson:"hardwarePlatform,omitempty"` // eg: x86_64
	OperatingSystem  string `bson:"operatingSystem,omitempty"`  // eg: GNU/Linux
}

type MachineMeta struct {
	Hostname   string             `bson:"hostname,omitempty"` // eg: ip-172-31-80-24
	Goos       string             `bson:"goos,omitempty"`     // eg: linux
	Goarch     string             `bson:"goarch,omitempty"`   // eg: amd64
	LsbRelease *LsbRelease        `bson:"lsbRelease,omitempty"`
	Uname      *Uname             `bson:"uname,omitempty"`
	AwsEc2Meta *AwsEc2IdentityDoc `bson:"awsEc2Meta,omitempty"`
}

type AwsEc2IdentityDoc struct {
	AccountId               string      `json:"accountId" bson:"accountId,omitempty"`
	Architecture            string      `json:"architecture" bson:"architecture,omitempty"`
	AvailabilityZone        string      `json:"availabilityZone" bson:"availabilityZone,omitempty"`
	BillingProducts         interface{} `json:"billingProducts" bson:"billingProducts,omitempty"`
	DevPayProductCodes      interface{} `json:"devpayProductCodes" bson:"devPayProductCodes,omitempty"` // The AWS's json field is 'devpayProductCodes', note the lowercase 'p' in 'pay'.
	MarketplaceProductCodes interface{} `json:"marketplaceProductCodes" bson:"marketplaceProductCodes,omitempty"`
	ImageId                 string      `json:"imageId" bson:"imageId,omitempty"`
	InstanceId              string      `json:"instanceId" bson:"instanceId,omitempty"`
	InstanceType            string      `json:"instanceType" bson:"instanceType,omitempty"`
	KernelId                interface{} `json:"kernelId" bson:"kernelId,omitempty"`
	PendingTime             time.Time   `json:"pendingTime" bson:"pendingTime,omitempty"`
	PrivateIp               string      `json:"privateIp" bson:"privateIp,omitempty"`
	RamdiskId               interface{} `json:"ramdiskId" bson:"ramdiskId,omitempty"`
	Region                  string      `json:"region" bson:"region,omitempty"`
	Version                 string      `json:"version" bson:"version,omitempty"`
}
