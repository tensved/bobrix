package ctx

const (
	BobrixCustomField   = "bobrix"
	AnswerToCustomField = BobrixCustomField + ".answer_to"
)

type MetadataKey struct{}

var MetadataKeyContext = MetadataKey{}
