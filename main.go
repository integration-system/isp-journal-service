package main

import (
	"context"
	"github.com/integration-system/isp-lib/v2/backend"
	"github.com/integration-system/isp-lib/v2/bootstrap"
	"github.com/integration-system/isp-lib/v2/config/schema"
	"github.com/integration-system/isp-lib/v2/structure"
	log "github.com/integration-system/isp-log"
	"github.com/integration-system/isp-log/stdcodes"
	"isp-journal-service/conf"
	"isp-journal-service/helper"
	"os"
)

var (
	version = "0.1.0"
)

func main() {
	bootstrap.
		ServiceBootstrap(&conf.Configuration{}, &conf.RemoteConfig{}).
		DefaultRemoteConfigPath(schema.ResolveDefaultConfigPath("default_remote_config.json")).
		OnLocalConfigLoad(onLocalConfigLoad).
		SocketConfiguration(socketConfiguration).
		OnSocketErrorReceive(onRemoteErrorReceive).
		OnConfigErrorReceive(onRemoteConfigErrorReceive).
		DeclareMe(makeDeclaration).
		OnRemoteConfigReceive(onRemoteConfigReceive).
		OnShutdown(onShutdown).
		Run()
}

func socketConfiguration(cfg interface{}) structure.SocketConfiguration {
	appConfig := cfg.(*conf.Configuration)
	return structure.SocketConfiguration{
		Host:   appConfig.ConfigServiceAddress.IP,
		Port:   appConfig.ConfigServiceAddress.Port,
		Secure: false,
		UrlParams: map[string]string{
			"module_name":   appConfig.ModuleName,
			"instance_uuid": appConfig.InstanceUuid,
		},
	}
}

func onShutdown(_ context.Context, _ os.Signal) {
	backend.StopGrpcServer()
}

func onRemoteConfigReceive(remoteConfig, oldConfig *conf.RemoteConfig) {

}

func onRemoteErrorReceive(errorMessage map[string]interface{}) {
	log.WithMetadata(errorMessage).Error(stdcodes.ReceiveErrorFromConfig, "error from config service")
}

func onRemoteConfigErrorReceive(errorMessage string) {
	log.WithMetadata(map[string]interface{}{
		"message": errorMessage,
	}).Error(stdcodes.ReceiveErrorOnGettingConfigFromConfig, "error on getting remote configuration")
}

func onLocalConfigLoad(cfg *conf.Configuration) {
	endpoints := helper.GetAllEndpoints(cfg.ModuleName)
	service := backend.NewDefaultService(endpoints)
	backend.StartBackendGrpcServer(cfg.GrpcInnerAddress, service)
}

func makeDeclaration(localConfig interface{}) bootstrap.ModuleInfo {
	cfg := localConfig.(*conf.Configuration)
	endpoints := helper.GetAllEndpoints(cfg.ModuleName)
	return bootstrap.ModuleInfo{
		ModuleName:       cfg.ModuleName,
		ModuleVersion:    version,
		GrpcOuterAddress: cfg.GrpcOuterAddress,
		Endpoints:        endpoints,
	}
}
