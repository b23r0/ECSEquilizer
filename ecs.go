package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	ecs "github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
)

type ECSMgr struct {
	AccessKeyId  string
	AccessSecret string
}

func (p *ECSMgr) create_ecs(region string, imageId string, instanceType string, securityGroupId string, internetMaxBandwidthIn int, vSwitchId string) (string, string, int) {
	client, err := ecs.NewClientWithAccessKey(region, p.AccessKeyId, p.AccessSecret)

	if err != nil {
		log.Panicln(err)
	}

	request := ecs.CreateCreateInstanceRequest()
	request.Scheme = "https"

	request.RegionId = region
	request.ImageId = imageId
	request.InstanceType = instanceType //"ecs.s6-c1m1.small"
	request.InternetChargeType = "PayByBandwidth"
	request.SecurityGroupId = securityGroupId                                    //"sg-bp1buct0j6jykdapgt4g"
	request.InternetMaxBandwidthIn = requests.NewInteger(internetMaxBandwidthIn) //requests.NewInteger(5)
	request.PasswordInherit = requests.NewBoolean(true)
	request.InstanceChargeType = "PostPaid"
	request.VSwitchId = vSwitchId
	request.InternetMaxBandwidthOut = requests.NewInteger(5)

	response, err := client.CreateInstance(request)
	if err != nil {
		log.Println(err.Error())
		return "", "", -1
	}
	id := response.InstanceId
	ip, err := p.allocate_public_ip(region, id)

	if err != nil {

		if strings.Contains(err.Error(), "IncorrectInstanceStatus") {
			log.Println("ECS intailizing , wait a moment.")
			// when ECS intailizing , allocate ip faild , wait to intailized . but just wait once.
			time.Sleep(5 * time.Second)
			ip, err = g_ecs.allocate_public_ip(region, id)
		}

		//if still faild
		if err != nil {
			log.Println(err.Error())
			return "", "", -2
		}
	}

	p.start_ecs(region, id)

	return ip, id, 0
}

func (p *ECSMgr) delete_ecs(region string, instanceId string) int {
	client, err := ecs.NewClientWithAccessKey(region, p.AccessKeyId, p.AccessSecret)

	if err != nil {
		log.Panicln(err)
	}

	request := ecs.CreateDeleteInstanceRequest()
	request.Scheme = "https"

	request.InstanceId = instanceId
	request.Force = requests.NewBoolean(true)

	_, err = client.DeleteInstance(request)
	if err != nil {
		log.Println(err.Error())
		if strings.Contains(err.Error(), "IncorrectInstanceStatus.Initializing") {
			return -1
		} else {
			return -2
		}
	}
	return 0
}

func (p *ECSMgr) start_ecs(region string, instanceId string) error {
	client, err := ecs.NewClientWithAccessKey(region, p.AccessKeyId, p.AccessSecret)

	if err != nil {
		log.Panicln(err)
	}

	request := ecs.CreateStartInstanceRequest()
	request.Scheme = "https"

	request.InstanceId = instanceId

	_, err = client.StartInstance(request)
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func (p *ECSMgr) allocate_public_ip(region string, instanceId string) (string, error) {
	client, err := ecs.NewClientWithAccessKey(region, p.AccessKeyId, p.AccessSecret)

	if err != nil {
		log.Panicln(err)
	}
	request := ecs.CreateAllocatePublicIpAddressRequest()
	request.Scheme = "https"

	request.InstanceId = instanceId

	response, err := client.AllocatePublicIpAddress(request)
	if err != nil {
		fmt.Print(err.Error())
	}

	return response.IpAddress, err
}