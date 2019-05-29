package cli

type alias = string

type list struct {
	UNKNOWN        alias
	BLUEGREEN      alias
	STANDARD_CHART alias
	MICROSERVICE   alias
}

var Command = &list{
	UNKNOWN:        "unknown",
	BLUEGREEN:      "bluegreen",
	STANDARD_CHART: "standard-chart",
	MICROSERVICE:   "microservice",
}

func DetermineCommand(command string) string {
	switch command {
	case Command.BLUEGREEN:
		return Command.BLUEGREEN
	case Command.STANDARD_CHART:
		return Command.STANDARD_CHART
	case Command.MICROSERVICE:
		return Command.MICROSERVICE
	default:
		return Command.UNKNOWN
	}
}
