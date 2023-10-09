/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package command

const (
	AppEmoji        = "🧙"
	ApplicationName = "pixa"
	RootPsName      = "root-ps"
	SourceID        = "github.com/snivilised/pixa"
)

func Execute() error {
	return (&Bootstrap{}).Root().Execute()
}
