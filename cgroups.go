package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/spf13/viper"
)

func createCgroupSpecFromConfig(cfg *viper.Viper) (*specs.LinuxResources, error) {

	cgroupSpec := &specs.LinuxResources{}
	memString := cfg.GetString(MEMORY_CONFIG_KEY)
	memLimit, err := parseMemoryLimit(memString)
	if err != nil {
		return nil, fmt.Errorf("could not parse configured memory limit: %v", err)
	}
	if memLimit != 0 {
		cgroupSpec.Memory = &specs.LinuxMemory{
			Limit: &memLimit,
		}
	}

	totalCPULimit := int64(cfg.GetInt(CPU_CONFIG_KEY))
	if totalCPULimit != 0 {
		numCPU := int64(runtime.NumCPU())
		oneSecondInMicros := uint64(1000000)
		cpuLimitAsDecimal := float64(totalCPULimit) / 100.0
		unitsPerSecondFloat := cpuLimitAsDecimal * float64(oneSecondInMicros) * float64(numCPU)
		unitsPerSecond := int64(unitsPerSecondFloat)

		cgroupSpec.CPU = &specs.LinuxCPU{
			Period: &oneSecondInMicros,
			Quota:  &unitsPerSecond,
			Cpus:   "0",
			Mems:   "0",
		}
	}
	return cgroupSpec, nil
}

// parseMemoryLimit takes the user provided configuration string for memory limits and returns it
// in bytes. It supports "g"/"G" and "m"/M" for gigabyte and megabyte
func parseMemoryLimit(memory string) (int64, error) {

	if memory == "0" {
		return 0, nil
	}

	memory = strings.ToUpper(memory)

	if strings.HasSuffix(memory, "G") {
		i, err := strconv.ParseInt(strings.TrimSuffix(memory, "G"), 10, 64)
		if err != nil {
			return i, err
		}
		return i * 1000000000, nil
	}

	if strings.HasSuffix(memory, "M") {
		i, err := strconv.ParseInt(strings.TrimSuffix(memory, "M"), 10, 64)
		if err != nil {
			return i, err
		}
		return i * 1000000, nil
	}

	return 1, fmt.Errorf("not a valid configuration for memory limits: %s", memory)
}

// enterCgroup will enter the calling process into a cgroup identified by `name/name-<PID>`
// with the resources specified in `spec`
func enterCgroup(name string, spec *specs.LinuxResources) error {

	staticPath := fmt.Sprintf("/%s/calm-%d", name, os.Getpid())

	control, err := cgroups.New(cgroups.V1, cgroups.StaticPath(staticPath), spec)
	if err != nil {
		return fmt.Errorf("could not create cgroup: %v", err)
	}

	err = control.Add(cgroups.Process{Pid: os.Getpid()})
	if err != nil {
		return fmt.Errorf("could not add process to cgroup control: %v", err)
	}

	//TODO: set release_agent and notify_on_release to rmdir script

	return nil
}
