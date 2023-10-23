package magick

type GeneralParameters struct {
	NoW     int
	Profile string
}

type FilterParameters struct {
	FolderRexEx string
	FolderGlob  string
	FilesRexEx  string
	FilesGlob   string
}
