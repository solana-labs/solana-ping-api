package main

type config struct {
	InfluxDB influxdb
	Solana   solana
}

type influxdb struct {
	URL      string `env:"INFLUX_URL"`
	Username string `env:"INFLUX_USER"`
	Password string `env:"INFLUX_PWD"`
	Database string `env:"INFLUX_DATABASE"`
}

type solana struct {
	URL        string `env:"SOLANA_URL,required"`
	Percentile uint16 `env:"SOLANA_FEE_PERCENTILE" envDefault:"0"`
}
