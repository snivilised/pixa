package i18n

// TODO: Should be updated to use url of the implementing project,
// so should not be left as arcadia.
const PixaSourceID = "github.com/snivilised/pixa"

type pixaTemplData struct{}

func (td pixaTemplData) SourceID() string {
	return PixaSourceID
}
