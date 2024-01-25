/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package command

import (
	"github.com/snivilised/cobrass/src/assistant"
	"github.com/snivilised/cobrass/src/store"
	xi18n "github.com/snivilised/extendio/i18n"
	"github.com/snivilised/extendio/xfs/storage"
	"github.com/snivilised/pixa/src/app/proxy/common"
	"github.com/snivilised/pixa/src/i18n"
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
	CascadeFamName    = "cascade-family"
)

func Execute() error {
	return (&Bootstrap{
		Vfs: storage.UseNativeFS(),
	}).Root().Execute()
}

func (b *Bootstrap) buildRootCommand(container *assistant.CobraContainer) {
	rootCommand := container.Root()
	paramSet := assistant.NewParamSet[common.RootParameterSet](rootCommand)

	// --sample (pending: sampling-family)
	//
	paramSet.BindBool(&assistant.FlagInfo{
		Name: "sample",
		Usage: i18n.LeadsWith(
			"sample",
			xi18n.Text(i18n.RootCmdSampleUsageTemplData{}),
		),
		Default:            false,
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	},
		&paramSet.Native.IsSampling,
	)

	const (
		defFSItems = uint(3)
		minFSItems = uint(1)
		maxFSItems = uint(128)
	)

	// --no-files (pending: sampling-family)
	//
	paramSet.BindValidatedUintWithin(
		&assistant.FlagInfo{
			Name: "no-files",
			Usage: i18n.LeadsWith(
				"no-files",
				xi18n.Text(i18n.RootCmdNoFilesUsageTemplData{}),
			),
			Default:            defFSItems,
			AlternativeFlagSet: rootCommand.PersistentFlags(),
		},
		&paramSet.Native.NoFiles,
		minFSItems,
		maxFSItems,
	)

	// --no-folders (pending: sampling-family)
	//
	paramSet.BindValidatedUintWithin(
		&assistant.FlagInfo{
			Name: "no-folders",
			Usage: i18n.LeadsWith(
				"no-folders",
				xi18n.Text(i18n.RootCmdNoFoldersUsageTemplData{}),
			),
			Default:            defFSItems,
			AlternativeFlagSet: rootCommand.PersistentFlags(),
		},
		&paramSet.Native.NoFolders,
		minFSItems,
		maxFSItems,
	)

	// --last (pending: sampling-family)
	//
	paramSet.BindBool(&assistant.FlagInfo{
		Name: "last",
		Usage: i18n.LeadsWith(
			"last",
			xi18n.Text(i18n.RootCmdLastUsageTemplData{}),
		),
		Default:            false,
		AlternativeFlagSet: rootCommand.PersistentFlags(),
	},
		&paramSet.Native.Last,
	)

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

	// family: cascade [--depth, --skim(K)]
	//
	cascadeFam := assistant.NewParamSet[store.CascadeParameterSet](rootCommand)
	cascadeFam.Native.BindAll(cascadeFam, rootCommand.PersistentFlags())

	// ??? rootCommand.Args = validatePositionalArgs

	container.MustRegisterParamSet(RootPsName, paramSet)
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
	}
}
