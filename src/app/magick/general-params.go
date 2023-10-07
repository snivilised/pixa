package magick

type GeneralParameters struct {
	Preview bool
	NoW     int
}

type FilterParameters struct {
	FolderRexEx string
	FolderGlob  string
	FilesRexEx  string
	FilesGlob   string
}
