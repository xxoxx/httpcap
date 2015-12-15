package config

type Flags struct {
	Verbose            bool
	InterfaceName      string
	Port               string
	Format             string
	Raw                bool
	Filter             string
	TruncateBodyLength int
        Service            string
}

var Setting Flags
