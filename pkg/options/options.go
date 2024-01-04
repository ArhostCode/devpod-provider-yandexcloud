package options

import (
	"fmt"
	"os"
	"strings"
)

var (
	YC_ZONE           = "YC_ZONE"
	YC_PLATFORM_ID    = "YC_PLATFORM_ID"
	YC_FOLDER_ID      = "YC_FOLDER_ID"
	YC_DISK_SIZE_GB   = "YC_DISK_SIZE_GB"
	MACHINE_ID        = "MACHINE_ID"
	MACHINE_FOLDER    = "MACHINE_FOLDER"
	YC_MEMORY_SIZE_GB = "YC_MEMORY_SIZE_GB"
	YC_CORES_COUNT    = "YC_CORES_COUNT"
)

type Options struct {
	Zone       string
	PlatformId string
	DiskSizeGB string
	FolderId   string
	CoresCount string
	RAMSizeGB  string

	MachineID     string
	MachineFolder string
}

func ConfigFromEnv() (Options, error) {
	return Options{
		Zone: os.Getenv(YC_ZONE),
	}, nil
}

func FromEnv(init bool) (*Options, error) {
	retOptions := &Options{}

	var err error

	retOptions.Zone, err = fromEnvOrError(YC_ZONE)
	if err != nil {
		return nil, err
	}
	retOptions.Zone = strings.ToLower(retOptions.Zone)

	retOptions.PlatformId, err = fromEnvOrError(YC_PLATFORM_ID)
	if err != nil {
		return nil, err
	}
	retOptions.DiskSizeGB, err = fromEnvOrError(YC_DISK_SIZE_GB)
	if err != nil {
		return nil, err
	}
	retOptions.FolderId, err = fromEnvOrError(YC_FOLDER_ID)
	if err != nil {
		return nil, err
	}
	retOptions.CoresCount, err = fromEnvOrError(YC_CORES_COUNT)
	if err != nil {
		return nil, err
	}
	retOptions.RAMSizeGB, err = fromEnvOrError(YC_MEMORY_SIZE_GB)
	if err != nil {
		return nil, err
	}

	// Return eraly if we're just doing init
	if init {
		return retOptions, nil
	}

	retOptions.MachineID, err = fromEnvOrError(MACHINE_ID)
	if err != nil {
		return nil, err
	}
	// prefix with devpod-
	retOptions.MachineID = "devpod-" + retOptions.MachineID

	retOptions.MachineFolder, err = fromEnvOrError(MACHINE_FOLDER)
	if err != nil {
		return nil, err
	}
	return retOptions, nil
}

func fromEnvOrError(name string) (string, error) {
	val := os.Getenv(name)
	if val == "" {
		return "", fmt.Errorf(
			"couldn't find option %s in environment, please make sure %s is defined",
			name,
			name,
		)
	}

	return val, nil
}
