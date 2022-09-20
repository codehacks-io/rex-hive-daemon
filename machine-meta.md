
wget -q -O - http://169.254.169.254/latest/dynamic/instance-identity/document

```json
{
  "accountId" : "463385481865",
  "architecture" : "x86_64",
  "availabilityZone" : "us-east-1b",
  "billingProducts" : null,
  "devpayProductCodes" : null,
  "marketplaceProductCodes" : null,
  "imageId" : "ami-04505e74c0741db8d",
  "instanceId" : "i-02fc2df1209ccfdb1",
  "instanceType" : "t2.micro",
  "kernelId" : null,
  "pendingTime" : "2022-09-14T22:34:41Z",
  "privateIp" : "172.31.17.42",
  "ramdiskId" : null,
  "region" : "us-east-1",
  "version" : "2017-09-30"
}
```

wget -q -O - http://169.254.169.254/latest/meta-data

```text
ami-id
ami-launch-index
ami-manifest-path
block-device-mapping/
events/
hibernation/
hostname
identity-credentials/
instance-action
instance-id
instance-life-cycle
instance-type
local-hostname
local-ipv4
mac
metrics/
network/
placement/
profile
public-hostname
public-ipv4
public-keys/
reservation-id
security-groups
services/
```

wget -q -O - http://169.254.169.254/latest/meta-data/ami-id
ami-04505e74c0741db8d

wget -q -O - http://169.254.169.254/latest/meta-data/ami-launch-index
0

wget -q -O - http://169.254.169.254/latest/meta-data/ami-manifest-path
(unknown)

wget -q -O - http://169.254.169.254/latest/meta-data/block-device-mapping/
wget -q -O - http://169.254.169.254/latest/meta-data/events/
wget -q -O - http://169.254.169.254/latest/meta-data/hibernation/
wget -q -O - http://169.254.169.254/latest/meta-data/hostname
ip-172-31-17-42.ec2.internal

wget -q -O - http://169.254.169.254/latest/meta-data/identity-credentials/
wget -q -O - http://169.254.169.254/latest/meta-data/instance-action
none

wget -q -O - http://169.254.169.254/latest/meta-data/instance-id
i-02fc2df1209ccfdb1

wget -q -O - http://169.254.169.254/latest/meta-data/instance-life-cycle
on-demand

wget -q -O - http://169.254.169.254/latest/meta-data/instance-type
t2.micro

wget -q -O - http://169.254.169.254/latest/meta-data/local-hostname
ip-172-31-17-42.ec2.internal

wget -q -O - http://169.254.169.254/latest/meta-data/local-ipv4
172.31.17.42

wget -q -O - http://169.254.169.254/latest/meta-data/mac
0a:00:19:73:6b:d3

wget -q -O - http://169.254.169.254/latest/meta-data/metrics/

wget -q -O - http://169.254.169.254/latest/meta-data/network/
wget -q -O - http://169.254.169.254/latest/meta-data/network/interfaces/macs/0a:00:19:73:6b:d3
wget -q -O - http://169.254.169.254/latest/meta-data/network/interfaces/macs/0a:00:19:73:6b:d3/owner-id
463385481865

wget -q -O - http://169.254.169.254/latest/meta-data/network/interfaces/macs/0a:00:19:73:6b:d3/public-hostname
ec2-54-242-223-204.compute-1.amazonaws.com

wget -q -O - http://169.254.169.254/latest/meta-data/placement/
wget -q -O - http://169.254.169.254/latest/meta-data/placement/availability-zone
us-east-1b
wget -q -O - http://169.254.169.254/latest/meta-data/placement/availability-zone-id
use1-az4

wget -q -O - http://169.254.169.254/latest/meta-data/profile
default-hvm

wget -q -O - http://169.254.169.254/latest/meta-data/public-hostname
ec2-54-242-223-204.compute-1.amazonaws.com

wget -q -O - http://169.254.169.254/latest/meta-data/public-ipv4
54.242.223.204

wget -q -O - http://169.254.169.254/latest/meta-data/public-keys/
0=the-name-of-my-ayuwoka-keyarda
wget -q -O - http://169.254.169.254/latest/meta-data/public-keys/0/openssh-key
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDb5MqIuJobMeO6V5ZQteTKFCLvaiDU2je6/wvlqdgDTwxov6qpuzLKhrYSeAaZC/SsagtxxGo5b7HwTofej2skeQBT75dbuVAL0gViofF3AoZ/zA59dIGfqiBRaxz6Bi2h3uy8Tf7J8tbK4rRQHEgLbNSMMiNveMA4ZnwaJo6iEyFAbCnH10gbNGa0AN1JLfqg1xvpfxqWe5p/aY/PdydTWK15d6L1iIotCovKRh3VTC0o1rtUFsR+rqytdnUpCJitKdAuau30cW0slA1YQdYwE3ir8T/RkcWHyIHI0RkejZEj0ZPUPw4DgpY8/hJs3oyfPwKj9vWh4sQdRSuxULRRpm3/EghLou1vBIlWOpGt0zIf1tMEHrPvjxwdDI9p5pNcTGqnYNiwn9uk973+f4blK95WtZvQifRLd0h5a3F1ZR87qFQscWyJ5LpNTYG8dyABiUBKGUMfWHF0OBxchbBnmSnr6eIuJRxkIfzsIyqW8ykH1hc4WomJglv2oLjKoDE= the-name-of-my-ayuwoka-keyarda
> That's my public key (my-meow-key.pub) content

wget -q -O - http://169.254.169.254/latest/meta-data/reservation-id
r-02b2359c41e24582e

wget -q -O - http://169.254.169.254/latest/meta-data/security-groups
please-allow-tls-to-my-example

wget -q -O - http://169.254.169.254/latest/meta-data/services/
wget -q -O - http://169.254.169.254/latest/meta-data/services/domain
amazonaws.com
wget -q -O - http://169.254.169.254/latest/meta-data/services/partition
aws
