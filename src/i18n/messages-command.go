package i18n

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type RootCmdShortDescTemplData struct {
	arcadiaTemplData
}

func (td RootCmdShortDescTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "root-command-short-description",
		Description: "short description for the root command",
		Other:       "A brief description of your application",
	}
}

type RootCmdLongDescTemplData struct {
	arcadiaTemplData
}

func (td RootCmdLongDescTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "root-command-long-description",
		Description: "long description for the root command",
		Other: `A longer description that spans multiple lines and likely contains
		examples and usage of using your application. For example:
		
		Cobra is a CLI library for Go that empowers applications.
		This application is a tool to generate the needed files
		to quickly create a Cobra application.`,
	}
}

type RootCmdConfigFileUsageTemplData struct {
	arcadiaTemplData
	ConfigFileName string
}

func (td RootCmdConfigFileUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "root-command-config-file-usage",
		Description: "root command config flag usage",
		Other:       "config file (default is $HOME/{{.ConfigFileName}}.yml)",
	}
}

type RootCmdLangUsageTemplData struct {
	arcadiaTemplData
}

func (td RootCmdLangUsageTemplData) Message() *i18n.Message {
	return &i18n.Message{
		ID:          "root-command-language-usage",
		Description: "root command lang usage",
		Other:       "'lang' defines the language according to IETF BCP 47",
	}
}
