package common

import (
	"fmt"
	"path/filepath"
)

const (
	appName = "pixa"
	yml     = "yml"
	org     = "snivilised"
)

type (
	commandDefs struct {
		Shrink string
	}

	pixaDefs struct {
		AppName            string
		Emoji              string
		SourceID           string
		ConfigTestFilename string
		ConfigType         string
		SubPath            string
	}

	thirdPartyDefs struct {
		Magick string
		Dummy  string
		Fake   string
	}

	configDefs struct {
		ConfigFilename string
	}

	loggingDefs struct {
		LogFilename string
	}

	defaultDefs struct {
		Config  configDefs
		Logging loggingDefs
	}

	environmentDefs struct {
		Home string
	}

	filingDefs struct {
		JournalExt    string
		Discriminator string // helps to identify files that should be filtered out
	}

	interactionDefs struct {
		Names struct {
			Discovery string
			Primary   string
		}
	}

	definitions struct {
		Pixa        pixaDefs
		ThirdParty  thirdPartyDefs
		Commands    commandDefs
		Defaults    defaultDefs
		Environment environmentDefs
		Filing      filingDefs
		Interaction interactionDefs
	}
)

var Definitions = definitions{
	Pixa: pixaDefs{
		AppName:            appName,
		Emoji:              "🧙",
		SourceID:           "github.com/snivilised/pixa",
		ConfigTestFilename: fmt.Sprintf("%v-test", appName),
		ConfigType:         yml,
		SubPath:            filepath.Join(org, appName),
	},
	ThirdParty: thirdPartyDefs{
		Magick: "magick",
		Dummy:  "dummy",
		Fake:   "fake",
	},
	Commands: commandDefs{
		Shrink: "shrink",
	},
	Defaults: defaultDefs{
		Config: configDefs{
			ConfigFilename: fmt.Sprintf("%v.%v", appName, yml),
		},
		Logging: loggingDefs{
			LogFilename: fmt.Sprintf("%v.log", appName),
		},
	},
	Environment: environmentDefs{
		Home: "PIXA_HOME",
	},
	Filing: filingDefs{
		JournalExt:    ".txt",
		Discriminator: ".$",
	},
	Interaction: interactionDefs{
		Names: struct {
			Discovery string
			Primary   string
		}{
			Discovery: "discover",
			Primary:   "primary",
		},
	},
}
