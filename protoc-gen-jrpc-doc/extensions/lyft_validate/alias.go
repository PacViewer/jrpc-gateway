package extensions

import evp "github.com/pacviewer/jrpc-gateway/protoc-gen-jrpc-doc/extensions/envoyproxy_validate"

// ValidateRule represents a single validator rule from the (validate.rules) method option extension.
type ValidateRule = evp.ValidateRule

// ValidateExtension contains the rules set by the (validate.rules) method option extension.
type ValidateExtension = evp.ValidateExtension
