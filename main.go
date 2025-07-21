// main
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/schollz/progressbar/v3"
)

// batch processing config
const (
	BatchSize   = 5000 // number of telemetry records per batch
	WorkerCount = 10   // number of worker goroutines
	BufferSize  = 50   // channel buffer size
)

// batch job structure
type TelemetryBatch struct {
	TripID       int32
	Records      []TripTelemetry
	BatchID      int
	TotalBatches int
}

// cli flags
type cliFlags struct {
	connStr  string
	migrate  bool
	showHelp bool
	platform string
	dataDir  string
}

// valid datalayers - as they are displayed
var datalayerSuggestions = []string{
	"postgresql",
}
var currentDatalayer = "postgresql"

// templates for filling out connection string
type (
	ConnectionStrParser func(connectionStr string, example string) (map[string]string, error)
	connStringTemplate  struct {
		validationFunc ConnectionStrParser
		exampleConnStr string
	}
)

var connectionTemplates = map[string]connStringTemplate{
	"postgresql": {
		validationFunc: ParsePostgresURL,
		exampleConnStr: "postgresql://<user>:<pass>@<localhost>:<port>/<db>?<setting=value>",
	},
}

// validation functions
func ValidateDatalayer(s string) error {
	if s == "" {
		return fmt.Errorf("select a datalayer")
	}
	for _, v := range datalayerSuggestions {
		if s == v {
			currentDatalayer = v
			return nil
		}
	}
	return fmt.Errorf("unsuported datalayer: %s", s)
}

func ValidateConnStr(s string) error {
	if s == "" {
		return errors.New("connection string cannot be empty")
	}
	template, ok := connectionTemplates[currentDatalayer]
	if !ok { // should never occur
		return fmt.Errorf("no template found for datalayer: %s", currentDatalayer)
	}
	_, err := template.validationFunc(s, template.exampleConnStr)
	return err
}

func ValidatePort(s string) error {
	if s == "" {
		return errors.New("you have to select a port number")
	}

	// try to lookup the port to validate it
	if _, err := net.LookupPort("tcp", s); err != nil {
		return fmt.Errorf("invalid port number '%s' (must be between 1-65535)", s)
	}

	// check if port is already in use
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", s))
	if err != nil {
		return fmt.Errorf("port %s is already in use", s)
	}
	listener.Close()

	return nil
}

func ValidateDataDir(d string) error {
	if d == "" {
		return errors.New("dataDir cannot be empty")
	}
	_, err := os.Stat(d)
	if err != nil {
		return fmt.Errorf("issue finding the data folder: %v", err)
	}
	_, err = os.Stat(filepath.Join(d, "/metaData.csv"))
	if err != nil {
		return fmt.Errorf("`metaData.csv` file not found in the provided folder: %v", err)
	}
	return nil
}

func parseFlags() cliFlags {
	flags := cliFlags{}

	// connection string
	flag.StringVar(
		&flags.platform,
		"platform",
		"",
		"Data platform to use as the data layer (e.g., postgresql)",
	)
	flag.StringVar(&flags.connStr, "connStr", "", "Connection string to the datalayer")
	flag.BoolVar(&flags.showHelp, "help", false, "Show help")
	flag.BoolVar(
		&flags.migrate,
		"migrate",
		false,
		"Migrate the orca db prior to launching orca. Will need to be run at least once to provision the store before use",
	)
	flag.StringVar(&flags.dataDir, "dataDir", "", "Location to the ZTBus Data")
	flag.Parse()

	return flags
}

func validateFlags(flags cliFlags) error {
	if flags.showHelp {
		return nil
	}

	if flags.platform == "" {
		return fmt.Errorf("a platform selection is required")
	}
	if err := ValidateDatalayer(flags.platform); err != nil {
		return fmt.Errorf("invalid platform: %w", err)
	}

	if err := ValidateConnStr(flags.connStr); err != nil {
		return fmt.Errorf("invalid connection string: %w", err)
	}

	if err := ValidateDataDir(flags.dataDir); err != nil {
		return fmt.Errorf("invalid dataDir: %w", err)
	}

	return nil
}

func telemetryWorker(
	ctx context.Context,
	pool *pgxpool.Pool,
	jobs <-chan TelemetryBatch,
	results chan<- error,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for batch := range jobs {
		err := func() error {
			conn, err := pool.Acquire(ctx)
			if err != nil {
				return fmt.Errorf("could not acquire connection: %v", err)
			}
			defer conn.Release()

			tx, err := conn.Begin(ctx)
			if err != nil {
				return fmt.Errorf("could not start transaction: %v", err)
			}
			defer tx.Rollback(ctx)
			qtx := New(tx).WithTx(tx)

			routeID, err := qtx.GetBusRouteIdFromTripId(ctx, batch.TripID)
			if err != nil {
				slog.Error("could not get bus route id", "error", err)
				return err
			}
			telemetryParams := make([]InsertTelemetryParams, len(batch.Records))

			for ii, telemRow := range batch.Records {
				telemetryParams[ii] = InsertTelemetryParams{
					TripID: batch.TripID,
					Time: pgtype.Timestamp{
						Time:  time.Unix(int64(telemRow.TimeUnix), 0),
						Valid: true,
					},
					ElectricPowerDemand: pgtype.Float4{
						Float32: float32(telemRow.ElectricPowerDemand),
						Valid:   true,
					},
					GnssAltitude: pgtype.Float4{
						Float32: float32(*telemRow.GnssAltitude),
						Valid:   telemRow.GnssAltitude != nil,
					},
					GnssCourse: pgtype.Float4{
						Float32: float32(*telemRow.GnssCourse),
						Valid:   telemRow.GnssCourse != nil,
					},
					GnssLatitude: pgtype.Float4{
						Float32: float32(*telemRow.GnssLatitude),
						Valid:   telemRow.GnssLatitude != nil,
					},
					GnssLongitude: pgtype.Float4{
						Float32: float32(*telemRow.GnssLongitude),
						Valid:   telemRow.GnssLongitude != nil,
					},
					ItcsNumberOfPassengers: pgtype.Int4{
						Int32: int32(*telemRow.ItcsNumberOfPassengers),
						Valid: telemRow.ItcsNumberOfPassengers != nil,
					},
					ItcsStopName: pgtype.Text{
						String: *telemRow.ItcsStopName,
						Valid:  telemRow.ItcsStopName != nil,
					},
					OdometryArticulationAngle: pgtype.Float4{
						Float32: float32(telemRow.OdometryArticulationAngle),
						Valid:   true,
					},
					OdometrySteeringAngle: pgtype.Float4{
						Float32: float32(telemRow.OdometrySteeringAngle),
						Valid:   true,
					},
					OdometryVehicleSpeed: pgtype.Float4{
						Float32: float32(telemRow.OdometryVehicleSpeed),
						Valid:   true,
					},
					OdometryWheelSpeedFl: pgtype.Float4{
						Float32: float32(telemRow.OdometryWheelSpeedFl),
						Valid:   true,
					},
					OdometryWheelSpeedFr: pgtype.Float4{
						Float32: float32(telemRow.OdometryWheelSpeedFr),
						Valid:   true,
					},
					OdometryWheelSpeedMl: pgtype.Float4{
						Float32: float32(telemRow.OdometryWheelSpeedMl),
						Valid:   true,
					},
					OdometryWheelSpeedMr: pgtype.Float4{
						Float32: float32(telemRow.OdometryWheelSpeedMr),
						Valid:   true,
					},
					OdometryWheelSpeedRl: pgtype.Float4{
						Float32: float32(telemRow.OdometryWheelSpeedRl),
						Valid:   true,
					},
					OdometryWheelSpeedRr: pgtype.Float4{
						Float32: float32(telemRow.OdometryWheelSpeedRr),
						Valid:   true,
					},
					StatusDoorIsOpen: pgtype.Bool{
						Bool:  telemRow.StatusDoorIsOpen,
						Valid: true,
					},
					StatusGridIsAvailable: pgtype.Bool{
						Bool:  telemRow.StatusGridIsAvailable,
						Valid: true,
					},
					StatusHaltBrakeIsActive: pgtype.Bool{
						Bool:  telemRow.StatusHaltBrakeIsActive,
						Valid: true,
					},
					StatusParkBrakeIsActive: pgtype.Bool{
						Bool:  telemRow.StatusParkBrakeIsActive,
						Valid: true,
					},
					TemperatureAmbient: pgtype.Float4{
						Float32: float32(telemRow.TemperatureAmbient),
						Valid:   true,
					},
					TractionBrakePressure: pgtype.Float4{
						Float32: float32(telemRow.TractionBrakePressure),
						Valid:   true,
					},
					TractionTractionForce: pgtype.Float4{
						Float32: float32(telemRow.TractionTractionForce),
						Valid:   true,
					},
					BusRouteID: routeID,
				}
			}

			count, err := qtx.InsertTelemetry(ctx, telemetryParams)
			if err != nil {
				return fmt.Errorf("error during COPY FROM: %v", err)
			}

			if err := tx.Commit(ctx); err != nil {
				return fmt.Errorf("could not commit transaction: %v", err)
			}

			slog.Debug("written results", "count", count)

			return nil
		}()

		if err != nil {
			results <- fmt.Errorf("batch %d/%d failed: %v", batch.BatchID, batch.TotalBatches, err)
		} else {
			results <- nil
			slog.Debug(
				"Completed telemetry batch",
				"batch", fmt.Sprintf("%d/%d", batch.BatchID, batch.TotalBatches),
				"records", len(batch.Records),
			)
		}
	}
}

// helper function to split telemetry data into batches
func createTelemetryBatches(tripID int32, telemetryData []TripTelemetry) []TelemetryBatch {
	var batches []TelemetryBatch
	totalBatches := (len(telemetryData) + BatchSize - 1) / BatchSize // ceiling division

	for i := 0; i < len(telemetryData); i += BatchSize {
		end := i + BatchSize
		end = min(end, len(telemetryData))

		batch := TelemetryBatch{
			TripID:       tripID,
			Records:      telemetryData[i:end],
			BatchID:      (i / BatchSize) + 1,
			TotalBatches: totalBatches,
		}
		batches = append(batches, batch)
	}

	return batches
}

func runCLI(flags cliFlags) error {
	if flags.showHelp {
		flag.Usage()
		return nil
	}

	// stdout logger
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// perform migrations if requested
	slog.Debug("premigration")
	if flags.migrate {
		slog.Debug("migrating datalayer")
		err := MigrateDatalayer(flags.platform, flags.connStr)
		if err != nil {
			slog.Error("could not migrate the datalayer, exiting", "error", err)
			return err
		}
	}
	slog.Debug("postmigration")
	slog.Debug("starting data load")

	metadata, err := ParseMetadataCSV(filepath.Join(flags.dataDir, "metaData.csv"))
	if err != nil {
		return fmt.Errorf("could not parse metadata CSV: %v", err)
	}

	// create connection pool for parallel processing
	ctx := context.Background()
	poolConfig, err := pgxpool.ParseConfig(flags.connStr)
	if err != nil {
		return fmt.Errorf("error parsing connection string: %v", err)
	}

	// configure pool settings for optimal performance
	poolConfig.MaxConns = int32(15) // workers + some buffer for main operations
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("error creating connection pool: %v", err)
	}
	defer pool.Close()

	// create a single connection for trip management (non-telemetry operations)
	tripConn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("error acquiring trip connection: %v", err)
	}
	defer tripConn.Release()

	// for each trip
	bar := progressbar.Default(int64(len(metadata)))
	for _, m := range metadata {
		bar.Add(1)

		// Start transaction for trip creation
		tx, err := tripConn.Begin(ctx)
		if err != nil {
			return fmt.Errorf("could not start the transaction: %v", err)
		}

		qtx := New(tx)

		// add bus
		busID, err := qtx.CreateBus(ctx, pgtype.Text{String: m.BusNumber, Valid: true})
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("could not create bus id: %v", err)
		}

		// add route
		routeID, err := qtx.CreateRoute(ctx, pgtype.Text{String: m.BusRoute, Valid: true})
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("could not create route id: %v", err)
		}

		// grab the trip info for this metadata
		tripID, err := qtx.CreateTrip(ctx, CreateTripParams{
			Name:    m.Name,
			BusID:   pgtype.Int4{Int32: busID, Valid: true},
			RouteID: pgtype.Int4{Int32: routeID, Valid: true},
			StartTime: pgtype.Timestamp{
				Time:  time.Unix(int64(m.StartTimeUnix), 0),
				Valid: true,
			},
			EndTime: pgtype.Timestamp{
				Time:  time.Unix(int64(m.EndTimeUnix), 0),
				Valid: true,
			},
			DrivenDistanceKm: pgtype.Float4{
				Float32: float32(m.DrivenDistance),
				Valid:   true,
			},
			EnergyConsumptionKWh: pgtype.Int4{
				Int32: int32(m.EnergyConsumption),
				Valid: true,
			},
			ItcsPassengersMean: pgtype.Float4{
				Float32: float32(m.ItcsNumberOfPassengersMean),
				Valid:   true,
			},
			ItcsPassengersMin: pgtype.Int4{
				Int32: int32(m.ItcsNumberOfPassengersMin),
				Valid: true,
			},
			ItcsPassengersMax: pgtype.Int4{
				Int32: int32(m.ItcsNumberOfPassengersMax),
				Valid: true,
			},
			GridAvailableMean: pgtype.Float4{
				Float32: float32(m.StatusGridIsAvailableMean),
				Valid:   true,
			},
			TemperatureMean: pgtype.Float4{
				Float32: float32(m.TemperatureAmbientMean),
				Valid:   true,
			},
			TemperatureMin: pgtype.Float4{
				Float32: float32(m.TemperatureAmbientMin),
				Valid:   true,
			},
			TemperatureMax: pgtype.Float4{
				Float32: float32(m.TemperatureAmbientMax),
				Valid:   true,
			},
		})
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("could not create trip: %v", err)
		}

		// Commit trip creation transaction
		err = tx.Commit(ctx)
		if err != nil {
			return fmt.Errorf("could not commit trip transaction: %v", err)
		}

		// Now handle telemetry data with parallel processing
		tripTelemetry, err := ParseTripTelemetryCSV(filepath.Join(flags.dataDir, m.Name+".csv"))
		if err != nil {
			return fmt.Errorf("could not parse telemetry CSV for trip %s: %v", m.Name, err)
		}

		if len(tripTelemetry) == 0 {
			slog.Debug("No telemetry data for trip", "trip", m.Name)
			continue
		}

		// Create batches for parallel processing
		batches := createTelemetryBatches(tripID, tripTelemetry)
		slog.Debug(
			"Processing telemetry data",
			"trip",
			m.Name,
			"total_records",
			len(tripTelemetry),
			"batches",
			len(batches),
		)

		// Set up worker pool for this trip's telemetry
		jobs := make(chan TelemetryBatch, BufferSize)
		results := make(chan error, len(batches))
		var wg sync.WaitGroup

		// start workers
		for range WorkerCount {
			wg.Add(1)
			go telemetryWorker(ctx, pool, jobs, results, &wg)
		}

		// send batches to workers
		go func() {
			defer close(jobs)
			for _, batch := range batches {
				jobs <- batch
			}
		}()

		// wait for all workers to finish
		go func() {
			wg.Wait()
			close(results)
		}()

		// collect results and check for errors
		var processingErrors []error
		for err := range results {
			if err != nil {
				processingErrors = append(processingErrors, err)
			}
		}

		if len(processingErrors) > 0 {
			return fmt.Errorf(
				"errors processing telemetry for trip %s: %v",
				m.Name,
				processingErrors[0],
			)
		}

		slog.Debug(
			"Successfully processed trip",
			"trip",
			m.Name,
			"telemetry_records",
			len(tripTelemetry),
		)
	}

	slog.Debug("Data load completed successfully")

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("could not acquire connection: %v", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not start transaction: %v", err)
	}
	defer tx.Rollback(ctx)
	qtx := New(tx).WithTx(tx)

	err = qtx.MakePartitions(ctx)
	if err != nil {
		return fmt.Errorf("could not create time partitions: %v", err)
	}
	slog.Debug("created partitions")

	return nil
}

func main() {
	flags := parseFlags()

	if err := validateFlags(flags); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	err := runCLI(flags)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}
