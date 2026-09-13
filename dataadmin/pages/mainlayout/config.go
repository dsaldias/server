package mainlayout

import "github.com/dsaldias/server/dataadmin/pages/mainlayout/layoutconfig"

type LayoutConfig = layoutconfig.Config

func DefaultLayoutConfig() LayoutConfig {
	return layoutconfig.Default()
}
