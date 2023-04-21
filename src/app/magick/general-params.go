package magick

type GeneralParameters struct {
	Preview bool
}

type FilterParameters struct {
	FolderRexEx string
	FolderGlob  string
	FilesRexEx  string
	FilesGlob   string
}
