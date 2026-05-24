package config

type Config struct {
	MaximumWeightDifference  string  `env:"MATCH_MAXIMUM_WEIGHT_DIFFERENCE"`
	LocationPollIntervalSecs int     `env:"LOCATION_POLL_INTERVAL_SECONDS" envDefault:"30"`
	NearbyRadiusKm           float64 `env:"NEARBY_RADIUS_KM" envDefault:"5"`
	RideRequestTTLMinutes    int     `env:"RIDE_REQUEST_TTL_MINUTES" envDefault:"15"`
}
