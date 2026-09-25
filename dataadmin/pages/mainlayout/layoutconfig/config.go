package layoutconfig

// mi clase de ayuda
type JSLink struct {
	Src  string
	Type string
}

type Config struct {
	Title        string
	ThemeColor   string
	Stylesheet   string
	BodyClass    string
	ShellClass   string
	MainClass    string
	SidebarClass string
	HeaderClass  string
	ExtraCSS     []string
	ExtraJS      []JSLink
}

func Default() Config {
	return Config{
		Title:        "Panel Admin",
		ThemeColor:   "#081027",
		Stylesheet:   "/assets/css/output.css",
		BodyClass:    "min-h-screen bg-background text-foreground",
		ShellClass:   "flex h-full min-h-0 flex-col",
		MainClass:    "min-h-0 flex-1 overflow-auto px-0 pt-0 pb-16",
		SidebarClass: "text-white z-[100] !border-r-0",
		HeaderClass:  "sticky top-0 flex h-10 w-full shrink-0 items-center justify-between border-b px-6 bg-header text-header-foreground z-[999]",
	}
}

func (c Config) WithDefaults() Config {
	defaults := Default()
	if c.Title == "" {
		c.Title = defaults.Title
	}
	if c.ThemeColor == "" {
		c.ThemeColor = defaults.ThemeColor
	}
	if c.Stylesheet == "" {
		c.Stylesheet = defaults.Stylesheet
	}
	if c.BodyClass == "" {
		c.BodyClass = defaults.BodyClass
	}
	if c.ShellClass == "" {
		c.ShellClass = defaults.ShellClass
	}
	if c.MainClass == "" {
		c.MainClass = defaults.MainClass
	}
	if c.SidebarClass == "" {
		c.SidebarClass = defaults.SidebarClass
	}
	if c.HeaderClass == "" {
		c.HeaderClass = defaults.HeaderClass
	}
	return c
}
