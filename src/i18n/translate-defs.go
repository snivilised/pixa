package i18n

// TODO: Should be updated to use url of the implementing project,
// so should not be left as arcadia.
const ArcadiaSourceID = "github.com/snivilised/arcadia"

type arcadiaTemplData struct{}

func (td arcadiaTemplData) SourceID() string {
	return ArcadiaSourceID
}
