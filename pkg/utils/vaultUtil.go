package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/trimble-oss/tierceron-core/pkg/core"
	helperkv "github.com/trimble-oss/tierceron-core/pkg/vaulthelper/kv"
	sys "github.com/trimble-oss/tierceron-core/pkg/vaulthelper/system"
)

// Helper to easiliy intialize a vault and a mod all at once.
func InitVaultMod(coreConfig *core.CoreConfig) (*core.CoreConfig, *helperkv.Modifier, *sys.Vault, error) {
	LogInfo(coreConfig, "InitVaultMod begins..")
	if coreConfig == nil {
		LogInfo(coreConfig, "InitVaultMod failure.  driverConfig provided is nil")
		return coreConfig, nil, nil, errors.New("invalid nil driverConfig")
	}

	vault, err := sys.NewVault(coreConfig.Insecure, coreConfig.VaultAddress, coreConfig.Env, false, false, false, coreConfig.Log)
	if err != nil {
		LogInfo(coreConfig, "Failure to connect to vault..")
		LogErrorObject(coreConfig, err, false)
		return coreConfig, nil, nil, err
	}
	vault.SetToken(coreConfig.Token)
	LogInfo(coreConfig, "InitVaultMod - Initializing Modifier")
	mod, err := helperkv.NewModifierFromCoreConfig(coreConfig, coreConfig.Env, false)
	if err != nil {
		LogErrorObject(coreConfig, err, false)
		return coreConfig, nil, nil, err
	}
	mod.Env = coreConfig.Env
	mod.Version = "0"
	mod.VersionFilter = coreConfig.VersionFilter
	LogInfo(coreConfig, "InitVaultMod complete..")

	return coreConfig, mod, vault, nil
}

var logMap sync.Map = sync.Map{}

// Helper to easiliy intialize a vault and a mod all at once.
func InitVaultModForPlugin(pluginConfig map[string]interface{}, logger *log.Logger) (*core.CoreConfig, *helperkv.Modifier, *sys.Vault, error) {
	logger.Println("InitVaultModForPlugin log setup: " + pluginConfig["env"].(string))
	var trcdbEnvLogger *log.Logger

	if _, nameSpaceOk := pluginConfig["logNamespace"]; nameSpaceOk {
		logPrefix := fmt.Sprintf("[trcplugin%s-%s]", pluginConfig["logNamespace"].(string), pluginConfig["env"].(string))

		if logger.Prefix() != logPrefix {
			logFile := fmt.Sprintf("/var/log/trcplugin%s-%s.log", pluginConfig["logNamespace"].(string), pluginConfig["env"].(string))
			if tLogger, logOk := logMap.Load(logFile); !logOk {
				logger.Printf("Checking log permissions for logfile: %s\n", logFile)

				f, logErr := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
				if logErr != nil {
					logFile = fmt.Sprintf("trcplugin%s-%s.log", pluginConfig["logNamespace"].(string), pluginConfig["env"].(string))
					f, logErr = os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
					if logErr != nil {
						logger.Println("Log permissions failure.  Will exit.")
					}
				}

				trcdbEnvLogger = log.New(f, fmt.Sprintf("[trcplugin%s-%s]", pluginConfig["logNamespace"].(string), pluginConfig["env"].(string)), log.LstdFlags)
				CheckError(&core.CoreConfig{ExitOnFailure: true, Log: trcdbEnvLogger}, logErr, true)
				logMap.Store(logFile, trcdbEnvLogger)
				logger.Println("InitVaultModForPlugin log setup complete")
			} else {
				logger.Printf("Utilizing existing logger for logfile: %s\n", logFile)
				trcdbEnvLogger = tLogger.(*log.Logger)
			}
		} else {
			trcdbEnvLogger = logger
		}
	} else {
		logger.Printf("Utilizing default logger invalid namespace\n")
		trcdbEnvLogger = logger
	}

	trcdbEnvLogger.Println("InitVaultModForPlugin begin..")
	exitOnFailure := false
	if _, ok := pluginConfig["exitOnFailure"]; ok {
		exitOnFailure = pluginConfig["exitOnFailure"].(bool)
	}

	trcdbEnvLogger.Println("InitVaultModForPlugin initialize DriverConfig.")

	var regions []string
	if _, regionsOk := pluginConfig["regions"]; regionsOk {
		regions = pluginConfig["regions"].([]string)
	}

	coreConfig := core.CoreConfig{
		WantCerts:     false,
		Insecure:      !exitOnFailure, // Plugin has exitOnFailure=false ...  always local, so this is ok...
		Token:         pluginConfig["token"].(string),
		VaultAddress:  pluginConfig["vaddress"].(string),
		Env:           pluginConfig["env"].(string),
		Regions:       regions,
		ExitOnFailure: exitOnFailure,
		Log:           trcdbEnvLogger,
	}
	trcdbEnvLogger.Println("InitVaultModForPlugin ends..")

	return InitVaultMod(&coreConfig)
}
