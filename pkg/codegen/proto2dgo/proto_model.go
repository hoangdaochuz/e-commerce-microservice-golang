package proto2dgo

type ProtoModel struct {
	Syntax       string
	ProtoPackage string
	GoPackage    string
	ImportPaths  []ImportModel
	Services     []ServiceModel
	Messages     []MessageModel
	Enums        []EnumModel
}

type ServiceModel struct {
	Name    string
	Methods []MethodModel
}

type MethodModel struct {
	Name         string
	RequestType  string
	ResponseType string
	ConstantName string
}

type MessageModel struct {
	MessageName string
	Fields      []FieldModel
}

type FieldModel struct {
	Type       string
	IsRepeat   bool
	Name       string
	Order      int
	IsOptional bool
}

type EnumModel struct {
	Name       string
	EnumFields []EnumField
}

type EnumField struct {
	Key   string
	Value int
}

type ImportMode string

const (
	NORMAL = "normal"
	PUBLIC = "public"
	WEAK   = "weak"
)

type ImportModel struct {
	Path string
	Mode ImportMode
}
