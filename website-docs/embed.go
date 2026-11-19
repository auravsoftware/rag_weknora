// Package websitedocs provides an empty embed.FS placeholder for product docs
// when the website-docs markdown directory is omitted from the repository.
package websitedocs

import "embed"

// Pages holds the embedded documentation filesystem (empty when website-docs
// content is omitted).
var Pages embed.FS
