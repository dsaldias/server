package assets

import "embed"

// FS contiene los recursos frontend incluidos en la librería. El consumidor
// puede sustituirlos durante el desarrollo generando sus propios assets.
//
//go:embed css js
var FS embed.FS
