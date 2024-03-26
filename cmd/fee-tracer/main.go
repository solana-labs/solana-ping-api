package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sort"
	"time"

	solana_client "github.com/blocto/solana-go-sdk/client"
	solana_common "github.com/blocto/solana-go-sdk/common"
	solana_rpc "github.com/blocto/solana-go-sdk/rpc"
	"github.com/caarlos0/env/v10"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// init log
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// init config
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatal().Err(err).Msg("Config")
	}

	log.Info().Msg("Start")

	for {
		func() {
			// init solana client
			solanaConn := solana_client.NewClient(cfg.Solana.URL)

			// get the fees
			var prioritizationFees solana_rpc.PrioritizationFees
			var err error

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			prioritizationFees, err = solanaConn.GetRecentPrioritizationFeesWithConfig(
				ctx,
				[]solana_common.PublicKey{},
				solana_rpc.GetRecentPrioritizationFeesConfig{
					Percentile: cfg.Solana.Percentile,
				},
			)
			if err != nil {
				log.Error().Err(err).Msg("Solana Client")
				return
			}

			sort.Slice(prioritizationFees, func(i, j int) bool {
				return prioritizationFees[i].Slot > prioritizationFees[j].Slot
			})
			if len(prioritizationFees) == 0 {
				log.Warn().Err(fmt.Errorf("got empty fees")).Msg("Solana Client")
				return
			}
			fee := prioritizationFees[0]
			if e := log.Debug(); e.Enabled() {
				e.
					Uint64("slot", fee.Slot).
					Uint16("percentile", cfg.Solana.Percentile).
					Uint64("fee", fee.PrioritizationFee).
					Msg("Datapoint")
			}

			// write the datapoint
			if cfg.InfluxDB.URL != "" {
				influxConn := influxdb2.NewClient(
					cfg.InfluxDB.URL,
					fmt.Sprintf("%v/%v", cfg.InfluxDB.Username, cfg.InfluxDB.Password),
				)
				defer influxConn.Close()

				writeAPI := influxConn.WriteAPI("anzaxyz", cfg.InfluxDB.Database)

				writeAPI.WritePoint(
					influxdb2.NewPoint("fees",
						map[string]string{"percentile": fmt.Sprintf("%v", cfg.Solana.Percentile)},
						map[string]interface{}{"slot": int64(fee.Slot), "fee": int64(fee.PrioritizationFee)},
						time.Now(),
					),
				)
			}

			time.Sleep(100 * time.Microsecond)
		}()
	}
}
