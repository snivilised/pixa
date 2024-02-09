/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package command

import (
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/cobrass/src/store"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/proxy/common"
)

const (
	RootPsName                = "root-ps"
	PreviewFamName            = "preview-family"
	WorkerPoolFamName         = "worker-pool-family"
	FoldersFamName            = "folders-family"
	ProfileFamName            = "profile-family"
	CascadeFamName            = "cascade-family"
	SamplingFamName           = "sampling-family"
	TextualInteractionFamName = "textual-family"
)

func Execute() error {
	return (&Bootstrap{
		Vfs: storage.UseNativeFS(),
	}).Root().Execute()
}

func (b *Bootstrap) buildRootCommand(container *assistant.CobraContainer) {
	rootCommand := container.Root()
	paramSet := assistant.NewParamSet[common.RootParameterSet](rootCommand)

	// family: sampling [--sample, --no-files, --no-folders, --last]
	//
	samplingFam := assistant.NewParamSet[store.SamplingParameterSet](rootCommand)
	samplingFam.Native.BindAll(samplingFam, rootCommand.PersistentFlags())

	// family: textual-interaction [--no-tui]
	//
	textualFam := assistant.NewParamSet[store.TextualInteractionParameterSet](rootCommand)
	textualFam.Native.BindAll(textualFam, rootCommand.PersistentFlags())

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

	// family: profile [--profile(P), --scheme(S)]
	//
	profileFam := assistant.NewParamSet[store.ProfileParameterSet](rootCommand)
	profileFam.Native.BindAll(profileFam, rootCommand.PersistentFlags())

	// family: cascade [--depth, --no-recurse(N)]
	//
	cascadeFam := assistant.NewParamSet[store.CascadeParameterSet](rootCommand)
	cascadeFam.Native.BindAll(cascadeFam, rootCommand.PersistentFlags())

	// ??? rootCommand.Args = validatePositionalArgs

	container.MustRegisterParamSet(RootPsName, paramSet)
	container.MustRegisterParamSet(SamplingFamName, samplingFam)
	container.MustRegisterParamSet(TextualInteractionFamName, textualFam)
	container.MustRegisterParamSet(PreviewFamName, previewFam)
	container.MustRegisterParamSet(WorkerPoolFamName, workerPoolFam)
	container.MustRegisterParamSet(FoldersFamName, foldersFam)
	container.MustRegisterParamSet(ProfileFamName, profileFam)
	container.MustRegisterParamSet(CascadeFamName, cascadeFam)
}

func (b *Bootstrap) getRootInputs() *common.RootCommandInputs {
	return &common.RootCommandInputs{
		ParamSet: b.Container.MustGetParamSet(
			RootPsName,
		).(*assistant.ParamSet[common.RootParameterSet]),
		SamplingFam: b.Container.MustGetParamSet(
			SamplingFamName,
		).(*assistant.ParamSet[store.SamplingParameterSet]),
		TextualFam: b.Container.MustGetParamSet(
			TextualInteractionFamName,
		).(*assistant.ParamSet[store.TextualInteractionParameterSet]),
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
		CascadeFam: b.Container.MustGetParamSet(
			CascadeFamName,
		).(*assistant.ParamSet[store.CascadeParameterSet]),
		Configs:      b.Configs,
		Presentation: &b.Presentation,
	}
}
