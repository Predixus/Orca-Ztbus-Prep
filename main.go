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
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
) // cli flags
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
		return fmt.Errorf("Select a datalayer")
	}
	for _, v := range datalayerSuggestions {
		if s == v {
			currentDatalayer = v
			return nil
		}
	}
	return fmt.Errorf("Unsuported datalayer: %s", s)
}

func ValidateConnStr(s string) error {
	if s == "" {
		return errors.New("Connection string cannot be empty")
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
		return errors.New("You have to select a port number")
	}

	// try to lookup the port to validate it
	if _, err := net.LookupPort("tcp", s); err != nil {
		return fmt.Errorf("Invalid port number '%s' (must be between 1-65535)", s)
	}

	// check if port is already in use
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", s))
	if err != nil {
		return fmt.Errorf("Port %s is already in use", s)
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

func runCLI(flags cliFlags) error {
	if flags.showHelp {
		flag.Usage()
		return nil
	}

	// stdout logger
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// perform migrations if requested
	slog.Info("premigration")
	if flags.migrate {
		slog.Info("migrating datalayer")
		err := MigrateDatalayer(flags.platform, flags.connStr)
		if err != nil {
			slog.Error("could not migrate the datalayer, exiting", "error", err)
			return err
		}
	}
	slog.Info("postmigration")
	slog.Info("starting data load")

	metadata, err := ParseMetadataCSV(filepath.Join(flags.dataDir, "metaData.csv"))

	// create a connection with the postgresql db
	ctx := context.Background()
	db, err := pgx.Connect(ctx, flags.connStr)
	if err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}
	queries := New(db)

	// for each trip
	for _, m := range metadata {
		tx, err := db.Begin(ctx)
		if err != nil {
			return fmt.Errorf("Could not start the transaction: %v", err)
		}
		defer tx.Rollback(ctx)

		qtx := queries.WithTx(tx)

		// add bus
		busId, err := qtx.CreateBus(ctx, pgtype.Text{String: m.BusNumber, Valid: true})
		if err != nil {
			return fmt.Errorf("could not create bus id: %v", err)
		}

		// add route
		routeId, err := qtx.CreateRoute(ctx, pgtype.Text{String: m.BusRoute, Valid: true})
		if err != nil {
			return fmt.Errorf("could not create route id: %v", err)
		}

		// grab the trip info for this metadata
		qtx.CreateTrip(ctx, CreateTripParams{
			Name:                 m.Name,
			BusID:                pgtype.Int4{Int32: busId, Valid: true},
			RouteID:              pgtype.Int4{Int32: routeId, Valid: true},
			StartTime:            pgtype.Timestamp{Time: time.Unix(m.StartTimeUnix)},
			EndTime:              pgtype.Timestamp{Time: time.Unix(m.EndTimeUnix)},
			DrivenDistanceKm:     m.DrivenDistance,
			EnergyConsumptionKWh: m.EnergyConsumption,
			ItcsPassengersMean:   m.ItcsNumberOfPassengersMean,
			ItcsPassengersMin:    m.ItcsNumberOfPassengersMin,
			ItcsPassengersMax:    m.ItcsNumberOfPassengersMax,
			GridAvailableMean:    m.StatusGridIsAvailableMean,
			TemperatureMean:      m.TemperatureAmbientMean,
			TemperatureMin:       m.TemperatureAmbientMin,
			TemperatureMax:       m.TemperatureAmbientMax,
		})

		// go through each row and gear up the
	}
	if err != nil {
		return fmt.Errorf("Could not parse metadata file: %v", err)
	}
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
