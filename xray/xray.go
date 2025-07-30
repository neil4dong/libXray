package xray

import (
	"errors"
	"os"
	"runtime/debug"
	"sync"

	"github.com/xtls/libxray/nodep"
	"github.com/xtls/xray-core/common/cmdarg"
	"github.com/xtls/xray-core/core"
	_ "github.com/xtls/xray-core/main/distro/all"
)

var (
	coreServer *core.Instance
	// int id
	instanceId int = 1
	// id -> core.Instance
	coreInstanceMap = make(map[int]*core.Instance)
	// 添加一个锁
	mu sync.Mutex
)

func StartXray(configPath string) (*core.Instance, error) {
	file := cmdarg.Arg{configPath}
	config, err := core.LoadConfig("json", file)
	if err != nil {
		return nil, err
	}

	server, err := core.New(config)
	if err != nil {
		return nil, err
	}

	return server, nil
}

func InitEnv(datDir string) {
	os.Setenv("xray.location.asset", datDir)
}

// Run Xray instance.
// datDir means the dir which geosite.dat and geoip.dat are in.
// configPath means the config.json file path.
func RunXray(datDir string, configPath string) (err error) {
	InitEnv(datDir)
	nodep.InitForceFree()
	coreServer, err = StartXray(configPath)
	if err != nil {
		return
	}

	if err = coreServer.Start(); err != nil {
		return
	}

	debug.FreeOSMemory()
	return nil
}

// Stop Xray instance.
func StopXray() error {
	if coreServer != nil {
		err := coreServer.Close()
		coreServer = nil
		if err != nil {
			return err
		}
	}
	return nil
}

// Xray's version
func XrayVersion() string {
	return core.Version()
}

func RunXrayReturnInstanceId(datDir string, configPath string) (int, error) {
	// 添加一个锁
	mu.Lock()
	defer mu.Unlock()

	InitEnv(datDir)
	nodep.InitForceFree()
	instance, err := StartXray(configPath)
	if err != nil {
		return 0, err
	}

	if err = instance.Start(); err != nil {
		return 0, err
	}

	usedInstanceId := instanceId
	coreInstanceMap[usedInstanceId] = instance
	instanceId++
	return usedInstanceId, nil
}

func StopXrayByInstanceId(instanceId int) error {

	// 添加一个锁
	mu.Lock()
	defer mu.Unlock()

	instance, ok := coreInstanceMap[instanceId]
	if !ok {
		return errors.New("instance not found")
	}

	if instance != nil {
		err := instance.Close()
		delete(coreInstanceMap, instanceId)
		if err != nil {
			return err
		}
	}
	return nil
}
