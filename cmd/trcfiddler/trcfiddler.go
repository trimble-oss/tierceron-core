package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/trimble-oss/tierceron-core/buildopts"
	"github.com/trimble-oss/tierceron-core/buildopts/coreopts"
	"github.com/trimble-oss/tierceron-core/buildopts/memonly"
	"github.com/trimble-oss/tierceron-core/buildopts/memprotectopts"
	trcvutils "github.com/trimble-oss/tierceron-core/pkg/core/util"
	eUtils "github.com/trimble-oss/tierceron-core/pkg/utils"
)

const configDir = "/.tierceron/config.yml"
const envContextPrefix = "envContext: "

// This is a controller program that can act as any command line utility.
// The swiss army knife of tierceron if you will.
func main() {
	if memonly.IsMemonly() {
		memprotectopts.MemProtectInit(nil)
	}
	buildopts.NewOptionsBuilder(buildopts.LoadOptions())
	coreopts.NewOptionsBuilder(coreopts.LoadOptions())
	fmt.Println("Version: " + "1.00")
	flagset := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagset.Usage = func() {
		fmt.Fprintf(flagset.Output(), "Usage of %s:\n", os.Args[0])
		flagset.PrintDefaults()
		fmt.Fprintf(flagset.Output(), "\nexample: trcfiddler\n")
	}
	envPtr := flagset.String("env", "", "Environment to be seeded") //If this is blank -> use context otherwise override context.
	tokenPtr := flagset.String("token", "", "Vault access token")
	addrPtr := flagset.String("addr", "", "API endpoint for the vault")
	regionPtr := flagset.String("region", "", "Region to be processed") //If this is blank -> use context otherwise override context.
	flagset.Bool("diff", false, "Diff files")

	flagset.Parse(os.Args[1:])
	if flagset.NFlag() == 0 {
		flagset.Usage()
		os.Exit(0)
	}
	logFile := "/var/log/trcfiddler.log"
	f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf(err.Error(), err)
	}
	logger := log.New(f, "[INIT]", log.LstdFlags)
	pluginConfig := map[string]interface{}{}
	pluginConfig["env"] = envPtr
	pluginConfig["vaddress"] = *addrPtr
	if tokenPtr != nil {
		pluginConfig["token"] = *tokenPtr
	}
	pluginConfig["ExitOnFailure"] = true
	if *regionPtr != "" {
		pluginConfig["regions"] = []string{*regionPtr}
	}
	coreConfig, mod, vault, err := eUtils.InitVaultModForPlugin(pluginConfig, logger)
	if err != nil {
		log.Fatalf(err.Error(), err)
	}

	properties, readErr := trcvutils.NewProperties(coreConfig, vault, mod, mod.Env, "trchelloworld", "Certify")

	fmt.Printf("%v\n", properties)

	if readErr != nil {
		logger.Fatalf(readErr.Error(), readErr)
	}
}
