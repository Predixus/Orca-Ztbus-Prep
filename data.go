package main

import (
	"encoding/csv"
	"io"
	"os"
	"strconv"
)

type Metadata struct {
	Name                       string
	BusNumber                  string
	StartTimeUnix              int
	EndTimeUnix                int
	DrivenDistance             float64
	BusRoute                   string
	EnergyConsumption          int
	ItcsNumberOfPassengersMean float64
	ItcsNumberOfPassengersMin  float64
	ItcsNumberOfPassengersMax  float64
	StatusGridIsAvailableMean  float64
	TemperatureAmbientMean     float64
	TemperatureAmbientMin      float64
	TemperatureAmbientMax      float64
}

type TripTelemetry struct {
	TimeUnix                  int
	ElectricPowerDemand       float64
	GnssAltitude              *float64
	GnssCourse                *float64
	GnssLatitude              *float64
	GnssLongitude             *float64
	ItcsBusRoute              string
	ItcsNumberOfPassengers    *int
	ItcsStopName              *string
	OdometryArticulationAngle float64
	OdometrySteeringAngle     float64
	OdometryVehicleSpeed      float64
	OdometryWheelSpeedFl      float64
	OdometryWheelSpeedFr      float64
	OdometryWheelSpeedMl      float64
	OdometryWheelSpeedMr      float64
	OdometryWheelSpeedRl      float64
	OdometryWheelSpeedRr      float64
	StatusDoorIsOpen          bool
	TatusGridIsAvailable      bool
	StatusHaltBrakeIsActive   bool
	StatusParkBrakeIsActive   bool
	TemperatureAmbient        float64
	TractionBrakePressure     float64
	TractionTractionForce     float64
}

func parseOptionalFloat(s string) (*float64, error) {
	if s == "" {
		return nil, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	return &f, err
}

func parseOptionalInt(s string) (*int, error) {
	if s == "" {
		return nil, nil
	}
	i, err := strconv.Atoi(s)
	return &i, err
}

func parseOptionalString(s string) (*string, error) {
	if s == "" {
		return nil, nil
	}
	return &s, nil
}

func ParseMetadataCSV(path string) ([]Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true

	headers, err := r.Read()
	if err != nil {
		return nil, err
	}

	index := make(map[string]int)
	for i, h := range headers {
		index[h] = i
	}

	var out []Metadata

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		get := func(col string) string { return row[index[col]] }

		parseF := func(s string) float64 {
			f, _ := strconv.ParseFloat(s, 64)
			return f
		}

		parseI := func(s string) int {
			i, _ := strconv.Atoi(s)
			return i
		}

		m := Metadata{
			Name:                       get("name"),
			BusNumber:                  get("busNumber"),
			StartTimeUnix:              parseI(get("startTime_unix")),
			EndTimeUnix:                parseI(get("endTime_unix")),
			DrivenDistance:             parseF(get("drivenDistance")),
			BusRoute:                   get("busRoute"),
			EnergyConsumption:          parseI(get("energyConsumption")),
			ItcsNumberOfPassengersMean: parseF(get("itcs_numberOfPassengers_mean")),
			ItcsNumberOfPassengersMin:  parseF(get("itcs_numberOfPassengers_min")),
			ItcsNumberOfPassengersMax:  parseF(get("itcs_numberOfPassengers_max")),
			StatusGridIsAvailableMean:  parseF(get("status_gridIsAvailable_mean")),
			TemperatureAmbientMean:     parseF(get("temperature_ambient_mean")),
			TemperatureAmbientMin:      parseF(get("temperature_ambient_min")),
			TemperatureAmbientMax:      parseF(get("temperature_ambient_max")),
		}

		out = append(out, m)
	}

	return out, nil
}

func ParseTripTelemetryCSV(path string) ([]TripTelemetry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.TrimLeadingSpace = true

	headers, err := r.Read()
	if err != nil {
		return nil, err
	}

	index := make(map[string]int)
	for i, h := range headers {
		index[h] = i
	}

	var out []TripTelemetry

	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		get := func(col string) string { return row[index[col]] }

		parseF := func(s string) float64 {
			f, _ := strconv.ParseFloat(s, 64)
			return f
		}

		parseI := func(s string) int {
			i, _ := strconv.Atoi(s)
			return i
		}

		parseB := func(s string) bool {
			b, _ := strconv.ParseBool(s)
			return b
		}

		trip := TripTelemetry{
			TimeUnix:                  parseI(get("time_unix")),
			ElectricPowerDemand:       parseF(get("electric_powerDemand")),
			GnssAltitude:              must(parseOptionalFloat(get("gnss_altitude"))),
			GnssCourse:                must(parseOptionalFloat(get("gnss_course"))),
			GnssLatitude:              must(parseOptionalFloat(get("gnss_latitude"))),
			GnssLongitude:             must(parseOptionalFloat(get("gnss_longitude"))),
			ItcsBusRoute:              get("itcs_busRoute"),
			ItcsNumberOfPassengers:    must(parseOptionalInt(get("itcs_numberOfPassengers"))),
			ItcsStopName:              must(parseOptionalString(get("itcs_stopName"))),
			OdometryArticulationAngle: parseF(get("odometry_articulationAngle")),
			OdometrySteeringAngle:     parseF(get("odometry_steeringAngle")),
			OdometryVehicleSpeed:      parseF(get("odometry_vehicleSpeed")),
			OdometryWheelSpeedFl:      parseF(get("odometry_wheelSpeed_fl")),
			OdometryWheelSpeedFr:      parseF(get("odometry_wheelSpeed_fr")),
			OdometryWheelSpeedMl:      parseF(get("odometry_wheelSpeed_ml")),
			OdometryWheelSpeedMr:      parseF(get("odometry_wheelSpeed_mr")),
			OdometryWheelSpeedRl:      parseF(get("odometry_wheelSpeed_rl")),
			OdometryWheelSpeedRr:      parseF(get("odometry_wheelSpeed_rr")),
			StatusDoorIsOpen:          parseB(get("status_doorIsOpen")),
			TatusGridIsAvailable:      parseB(get("tatus_gridIsAvailable")),
			StatusHaltBrakeIsActive:   parseB(get("status_haltBrakeIsActive")),
			StatusParkBrakeIsActive:   parseB(get("status_parkBrakeIsActive")),
			TemperatureAmbient:        parseF(get("temperature_ambient")),
			TractionBrakePressure:     parseF(get("traction_brakePressure")),
			TractionTractionForce:     parseF(get("traction_tractionForce")),
		}

		out = append(out, trip)
	}

	return out, nil
}

func must[T any](v T, _ error) T { return v }
