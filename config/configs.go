package config

import (
	"net"
	"os"
	"stateful-service/slog"
	"stateful-service/utils"
	"strconv"
)

const (
	EnvRunningPort         = "RUNNING_PORT"
	EnvIdGeneratorHost     = "ID_GENERATOR_HOST"
	EnvCheckpointInterval  = "CHECKPOINT_INTERVAL"
	ControllerSourceString = "CONTROLLER"
	EnvCheckpointMode      = "CHECKPOINT_MODE"
)

func GetIp() (net.IP, int32) {
	netInterfaces, err := net.Interfaces()
	if err != nil {
		slog.Fatalf("net.Interfaces failed, err: %v", err.Error())
	}
	for i := 0; i < len(netInterfaces); i++ {
		if (netInterfaces[i].Flags & net.FlagUp) != 0 {
			addrs, _ := netInterfaces[i].Addrs()
			for _, address := range addrs {
				if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						ip, err := utils.IpToInt(ipnet.IP.To4())
						if err != nil {
							slog.Fatal(err)
						}
						slog.Infof("get ip %v, convert to %v", ipnet.IP.To4(), ip)
						return ipnet.IP.To4(), ip
					}
				}
			}
		}
	}
	return nil, 0
}

func GetIntEnv(envName string) int {
	envValue, err := strconv.Atoi(os.Getenv(envName))
	if err != nil {
		slog.Fatalf("get int env %s failed, err: %v", envName, err.Error())
	}
	return envValue
}

func GetBoolEnv(envName string) bool {
	envValue, err := strconv.ParseBool(os.Getenv(envName))
	if err != nil {
		slog.Fatalf("get bool env %s failed, err: %v", envName, err.Error())
	}
	return envValue
}

func GetFloatEnv(envName string, bitSizes int) float64 {
	envValue, err := strconv.ParseFloat(os.Getenv(envName), bitSizes)
	if err != nil {
		slog.Fatalf("get float64 env %s failed, err: %v", envName, err.Error())
	}
	return envValue
}

func GetStringEnv(envName string) string {
	envValue := os.Getenv(envName)
	if envValue == "" {
		slog.Fatalf("get string env %s failed, err: get empty string", envName)
	}
	return envValue
}
