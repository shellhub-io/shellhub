package request

// TagParam is a structure to represent and validate a tag as path param.
type TagParam struct {
	Tag string `param:"tag" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

// TagBody is a structure to represent and validate a tag as json request body.
type TagBody struct {
	Tag string `json:"tag" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}

// TagDelete is the structure to represent the request data for delete tag endpoint.
type TagDelete struct {
	TagParam
}

// TagRename is the structure to represent the request data for rename tag endpoint.
type TagRename struct {
	TagParam
	NewTag string `json:"tag" validate:"required,min=3,max=255,alphanum,ascii,excludes=/@&:"`
}
