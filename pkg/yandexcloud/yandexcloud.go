package yandexcloud

import (
	"context"
	"fmt"
	"github.com/ArhostCode/devpod-provider-yandexcloud/pkg/options"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/operation"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/vpc/v1"

	//v2 "github.com/exoscale/egoscale/v2"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/loft-sh/devpod/pkg/client"
	"github.com/loft-sh/log"
	"github.com/pkg/errors"
	ycapi "github.com/yandex-cloud/go-sdk"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type YCProvider struct {
	Config           *options.Options
	SDK              *ycapi.SDK
	Log              log.Logger
	WorkingDirectory string
}

func StringPtr(v string) *string {
	return &v
}

type defaultTransport struct {
	next http.RoundTripper
}

var userAgent string

func (t *defaultTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", userAgent)

	resp, err := t.next.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func NewProvider(logs log.Logger, init bool) (*YCProvider, error) {
	config, err := options.FromEnv(init)

	if err != nil {
		return nil, err
	}

	httpClient := cleanhttp.DefaultPooledClient()
	httpClient.Transport = &defaultTransport{next: httpClient.Transport}

	apiKey := os.Getenv("YC_API_KEY")
	if apiKey == "" {
		return nil, errors.Errorf("YC_API_KEY is not set")
	}
	//y0_AgAAAAA-z7NpAATuwQAAAAD3EpZdVHcRVJKFRJe7MJujZCcrw3NuaKo
	apiSecret := os.Getenv("YC_API_SECRET")
	if apiSecret == "" {
		return nil, errors.Errorf("YC_API_SECRET is not set")
	}

	ctx := context.Background()
	sdk, err := ycapi.Build(ctx, ycapi.Config{
		Credentials: ycapi.OAuthToken(apiKey),
	})
	if err != nil {
		return nil, err
	}
	return &YCProvider{
		SDK:    sdk,
		Log:    logs,
		Config: config,
	}, nil
}

func GetDevpodInstance(ctx context.Context, ycProvider *YCProvider) (*compute.Instance, error) {

	request := compute.ListInstancesRequest{
		FolderId: ycProvider.Config.FolderId,
		PageSize: 100,
	}

	instances, err := ycProvider.SDK.Compute().Instance().List(ctx, &request)
	if err != nil {
		return nil, err
	}
	var instanceID string = ""
	for _, instance := range instances.Instances {
		if strings.Contains(instance.Name, ycProvider.Config.MachineID) {
			ycProvider.Log.Debugf("Found instance %v\n", instance.Name)
			instanceID = instance.Id
			break
		}
	}
	if instanceID == "" {
		return nil, fmt.Errorf("instance not found")
	}

	getInstanceRequest := compute.GetInstanceRequest{
		InstanceId: instanceID,
	}

	instance, err := ycProvider.SDK.Compute().Instance().Get(ctx, &getInstanceRequest)
	if err != nil {
		return nil, fmt.Errorf("get instance: %w", err)
	}
	return instance, nil
}

func Init(ctx context.Context, exoscaleProvider *YCProvider) error {
	//ctx2 := exoapi.WithEndpoint(ctx, exoapi.NewReqEndpoint("", exoscaleProvider.Config.Zone))
	//_, err := exoscaleProvider.SDK.ListZones(ctx2)
	//if err != nil {
	//	return err
	//}
	return nil
}

func Create(ctx context.Context, ycProvider *YCProvider) error {
	_, err := createInstance(ctx, ycProvider)
	if err != nil {
		return err
	}
	return nil
}

func Delete(ctx context.Context, ycProvider *YCProvider) error {
	devPodInstance, err := GetDevpodInstance(ctx, ycProvider)
	if err != nil {
		return err
	}

	request := compute.DeleteInstanceRequest{
		InstanceId: devPodInstance.Id,
	}

	_, err = ycProvider.SDK.Compute().Instance().Delete(ctx, &request)

	if err != nil {
		return err
	}

	return nil
}

func Start(ctx context.Context, ycProvider *YCProvider) error {
	devPodInstance, err := GetDevpodInstance(ctx, ycProvider)
	if err != nil {
		return err
	}

	request := compute.StartInstanceRequest{
		InstanceId: devPodInstance.Id,
	}

	_, err2 := ycProvider.SDK.Compute().Instance().Start(ctx, &request)
	if err2 != nil {
		return err2
	}

	return nil
}

func Status(ctx context.Context, ycProvider *YCProvider) (client.Status, error) {
	devPodInstance, err := GetDevpodInstance(ctx, ycProvider)
	if err != nil {
		return client.StatusNotFound, nil
	}
	switch {
	case devPodInstance.Status == compute.Instance_RUNNING:
		return client.StatusRunning, nil
	case devPodInstance.Status == compute.Instance_STOPPED:
		return client.StatusStopped, nil
	default:
		return client.StatusBusy, nil
	}
}

func Stop(ctx context.Context, ycProvider *YCProvider) error {
	devPodInstance, err := GetDevpodInstance(ctx, ycProvider)
	if err != nil {
		return err
	}

	request := compute.StopInstanceRequest{
		InstanceId: devPodInstance.Id,
	}

	_, err = ycProvider.SDK.Compute().Instance().Stop(ctx, &request)
	if err != nil {
		return err
	}
	return nil
}

func createInstance(ctx context.Context, ycProvider *YCProvider) (*operation.Operation, error) {

	userData := fmt.Sprintf(`#cloud-config
users:
- name: devpod
  shell: /bin/bash
  groups: [ sudo, docker ]
  ssh_authorized_keys:
  - %s
  sudo: [ "ALL=(ALL) NOPASSWD:ALL" ]`, ycProvider.Config.SSHPublicKey)

	subnetID := findSubnet(ctx, ycProvider.SDK, ycProvider.Config.FolderId, ycProvider.Config.Zone)
	sourceImageID := sourceImage(ctx, ycProvider.SDK)

	bootDiskSize, err := strconv.Atoi(ycProvider.Config.DiskSizeGB)

	request := &compute.CreateInstanceRequest{
		FolderId:   ycProvider.Config.FolderId,
		Name:       ycProvider.Config.MachineID,
		ZoneId:     ycProvider.Config.Zone,
		PlatformId: "standard-v1",
		ResourcesSpec: &compute.ResourcesSpec{
			Cores:  1,
			Memory: 2 * 1024 * 1024 * 1024,
		},
		BootDiskSpec: &compute.AttachedDiskSpec{
			AutoDelete: true,
			Disk: &compute.AttachedDiskSpec_DiskSpec_{
				DiskSpec: &compute.AttachedDiskSpec_DiskSpec{
					TypeId: "network-hdd",
					Size:   int64(bootDiskSize * 1024 * 1024 * 1024),
					Source: &compute.AttachedDiskSpec_DiskSpec_ImageId{
						ImageId: sourceImageID,
					},
				},
			},
		},
		NetworkInterfaceSpecs: []*compute.NetworkInterfaceSpec{
			{
				SubnetId: subnetID,
				PrimaryV4AddressSpec: &compute.PrimaryAddressSpec{
					OneToOneNatSpec: &compute.OneToOneNatSpec{
						IpVersion: compute.IpVersion_IPV4,
					},
				},
			},
		},
		Metadata: map[string]string{"user-data": userData},
	}
	op, err := ycProvider.SDK.Compute().Instance().Create(ctx, request)
	return op, err
}

func findSubnet(ctx context.Context, sdk *ycapi.SDK, folderID string, zone string) string {
	resp, err := sdk.VPC().Subnet().List(ctx, &vpc.ListSubnetsRequest{
		FolderId: folderID,
		PageSize: 100,
	})
	if err != nil {
		return ""
	}
	subnetID := ""
	for _, subnet := range resp.Subnets {
		if subnet.ZoneId != zone {
			continue
		}
		subnetID = subnet.Id
		break
	}
	if subnetID == "" {
		return ""
	}
	return subnetID
}

func sourceImage(ctx context.Context, sdk *ycapi.SDK) string {
	image, err := sdk.Compute().Image().GetLatestByFamily(ctx, &compute.GetImageLatestByFamilyRequest{
		FolderId: "standard-images",
		Family:   "ubuntu-2204-lts",
	})
	if err != nil {
		return ""
	}
	return image.Id
}
