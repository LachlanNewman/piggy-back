package config

type Config struct {
	LogLevel                 string  `env:"LOG_LEVEL" envDefault:"info"`
	OIDCIssuer               string  `env:"OIDC_ISSUER"`
	OIDCAudience             string  `env:"OIDC_AUDIENCE"`
	MaximumWeightDifference  string  `env:"MATCH_MAXIMUM_WEIGHT_DIFFERENCE"`
	LocationPollIntervalSecs int     `env:"LOCATION_POLL_INTERVAL_SECONDS" envDefault:"30"`
	NearbyRadiusKm           float64 `env:"NEARBY_RADIUS_KM" envDefault:"5"`
	MaxNearbyRadiusKm        float64 `env:"MAX_NEARBY_RADIUS_KM" envDefault:"20"`
	RideRequestTTLMinutes    int     `env:"RIDE_REQUEST_TTL_MINUTES" envDefault:"15"`
}
