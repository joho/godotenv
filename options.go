package godotenv

// GoDotEnvOptions stores options for the parser
type GoDotEnvOptions struct {
	EscapeExclamation bool
}

// default options
var Options = GoDotEnvOptions{
	EscapeExclamation: true,
}
