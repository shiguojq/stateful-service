package config

import (
	"fmt"
	"testing"
)

func TestGetIp(t *testing.T) {
	ip, ipNum := GetIp()
	fmt.Println(ip, ipNum)
}

func TestGetRunningPort(t *testing.T) {
	runningPort := GetIntEnv("RUNNING_PORT")
	fmt.Println(runningPort)
}
