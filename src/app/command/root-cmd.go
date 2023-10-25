/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package command

import (
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/cobrass/src/store"
	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/pixa/src/app/proxy"
	"github.com/snivilised/pixa/src/i18n"
	"github.com/spf13/pflag"
	"golang.org/x/text/language"
)

const (
	AppEmoji          = "🧙"
	ApplicationName   = "pixa"
	SourceID          = "github.com/snivilised/pixa"
	RootPsName        = "root-ps"
	PreviewFamName    = "preview-family"
	WorkerPoolFamName = "worker-pool-family"
	FoldersFamName    = "folders-family"
	ProfileFamName    = "profile-family"
)

func Execute() error {
	return (&Bootstrap{}).Root().Execute()
}

func (b *Bootstrap) buildRootCommand(container *assistant.CobraContainer) {
	rootCommand := container.Root()
	paramSet := assistant.NewParamSet[proxy.RootParameterSet](rootCommand)

	// --lang (TODO: should really come from the family store,
	// as its a generic concept.)
	//
	paramSet.BindValidatedString(&assistant.FlagInfo{
		Name: "lang",
		Usage: i18n.LeadsWith(
			"lang",
			xi18n.Text(i18n.RootCmdLangUsageTemplData{}),
		),
		Default:            xi18n.DefaultLanguage.Get().String(),
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	}, &paramSet.Native.Language, func(value string, _ *pflag.Flag) error {
		_, err := language.Parse(value)
		return err
	})

	// family: preview [--dry-run(D)]
	//
	previewFam := assistant.NewParamSet[store.PreviewParameterSet](rootCommand)
	previewFam.Native.BindAll(previewFam, rootCommand.PersistentFlags())

	// family: worker-pool [--cpu(C), --now(N)]
	//
	workerPoolFam := assistant.NewParamSet[store.WorkerPoolParameterSet](rootCommand)
	workerPoolFam.Native.BindAll(workerPoolFam, rootCommand.PersistentFlags())

	// family: folders filter [--folders-gb(Z), --folders-rx(Y)]
	//
	// For now, we don't want the folders flags to be inherited by
	// the shrink command because the shrink command needs to be
	// able to define a different filter for files and folders
	// which needs to be implemented by a custom filter, which has
	// not been implemented yet. This is why we don't pass in
	// rootCommand.PersistentFlags() into BindAll below.
	foldersFam := assistant.NewParamSet[store.FoldersFilterParameterSet](rootCommand)
	foldersFam.Native.BindAll(foldersFam)

	// family: profile [--profile(p)]
	//
	profileFam := assistant.NewParamSet[store.ProfileParameterSet](rootCommand)
	profileFam.Native.BindAll(profileFam, rootCommand.PersistentFlags())

	rootCommand.Args = validatePositionalArgs

	container.MustRegisterParamSet(RootPsName, paramSet)
	container.MustRegisterParamSet(PreviewFamName, previewFam)
	container.MustRegisterParamSet(WorkerPoolFamName, workerPoolFam)
	container.MustRegisterParamSet(FoldersFamName, foldersFam)
	container.MustRegisterParamSet(ProfileFamName, profileFam)
}

func (b *Bootstrap) getRootInputs() *proxy.RootCommandInputs {
	return &proxy.RootCommandInputs{
		ParamSet: b.Container.MustGetParamSet(
			RootPsName,
		).(*assistant.ParamSet[proxy.RootParameterSet]),
		PreviewFam: b.Container.MustGetParamSet(
			PreviewFamName,
		).(*assistant.ParamSet[store.PreviewParameterSet]),
		WorkerPoolFam: b.Container.MustGetParamSet(
			WorkerPoolFamName,
		).(*assistant.ParamSet[store.WorkerPoolParameterSet]),
		FoldersFam: b.Container.MustGetParamSet(
			FoldersFamName,
		).(*assistant.ParamSet[store.FoldersFilterParameterSet]),
		ProfileFam: b.Container.MustGetParamSet(
			ProfileFamName,
		).(*assistant.ParamSet[store.ProfileParameterSet]),
	}
}
