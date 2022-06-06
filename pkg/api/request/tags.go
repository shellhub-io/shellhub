package request

// TagParam is a parameter that can be used to validate a tag.
type TagParam struct {
	// Tag's name.
	Tag string `param:"tag" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

// TagDelete is the structure for the request data at delete tag endpoint.
type TagDelete struct {
	TagParam
}

// TagRename is the structure for the request data at rename tag endpoint.
type TagRename struct {
	TagParam
	NewTag string `json:"tag" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}
