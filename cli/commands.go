package cli

type alias = string

type list struct {
	UNKNOWN   alias
	BLUEGREEN alias
	STANDARD  alias
}

var Command = &list{
	UNKNOWN:   "unknown",
	BLUEGREEN: "bluegreen",
	STANDARD:  "standard",
}

func DetermineCommand(command string) string {
	switch command {
	case Command.BLUEGREEN:
		return Command.BLUEGREEN
	case Command.STANDARD:
		return Command.STANDARD
	default:
		return Command.UNKNOWN
	}
}
