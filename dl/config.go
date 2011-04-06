package dl

type Config struct {
	Alpha, Kappa float64
	M            uint64
}

func ConfigDefault() (cfg Config) {
	cfg.Alpha = 1
	cfg.Kappa = 0.5
	cfg.M = 30
	return
}
