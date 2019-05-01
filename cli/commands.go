package cli

type alias = string

type list struct {
	UNKNOWN        alias
	BLUEGREEN      alias
	STANDARD_CHART alias
}

var Command = &list{
	UNKNOWN:        "unknown",
	BLUEGREEN:      "bluegreen",
	STANDARD_CHART: "standard-chart",
}

func DetermineCommand(command string) string {
	switch command {
	case Command.BLUEGREEN:
		return Command.BLUEGREEN
	case Command.STANDARD_CHART:
		return Command.STANDARD_CHART
	default:
		return Command.UNKNOWN
	}
}
