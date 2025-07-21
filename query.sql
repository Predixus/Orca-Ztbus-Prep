-- name: CreateBus :one
INSERT INTO buses (bus_number)
VALUES (sqlc.arg('bus_number'))
ON CONFLICT (bus_number) DO UPDATE SET bus_number = EXCLUDED.bus_number
RETURNING id;

-- name: CreateRoute :one
INSERT INTO bus_routes (route_code)
VALUES (sqlc.arg('route_code'))
ON CONFLICT (route_code) DO UPDATE SET route_code = EXCLUDED.route_code
RETURNING id;

-- name: CreateTrip :one
INSERT INTO trips (
  name,
  bus_id,
  route_id,
  start_time,
  end_time,
  driven_distance_km,
  energy_consumption_kWh,
  itcs_passengers_mean,
  itcs_passengers_min,
  itcs_passengers_max,
  grid_available_mean,
  amb_temperature_mean,
  amb_temperature_min,
  amb_temperature_max
)
VALUES (
  sqlc.arg('name'),
  sqlc.arg('bus_id'),
  sqlc.arg('route_id'),
  sqlc.arg('start_time'),
  sqlc.arg('end_time'),
  sqlc.arg('driven_distance_km'),
  sqlc.arg('energy_consumption_kWh'),
  sqlc.arg('itcs_passengers_mean'),
  sqlc.arg('itcs_passengers_min'),
  sqlc.arg('itcs_passengers_max'),
  sqlc.arg('grid_available_mean'),
  sqlc.arg('temperature_mean'),
  sqlc.arg('temperature_min'),
  sqlc.arg('temperature_max')
)
RETURNING id;

-- name: GetTripByName :one
SELECT * FROM trips
WHERE name = sqlc.arg('name');

-- name: GetTripsByBus :many
SELECT * FROM trips
WHERE bus_id = sqlc.arg('bus_id')
ORDER BY start_time;

-- name: GetTripsByRoute :many
SELECT * FROM trips
WHERE route_id = sqlc.arg('route_id')
ORDER BY start_time;

-- name: GetTripsByTimeRange :many
SELECT * FROM trips
WHERE start_time >= sqlc.arg('start_time_from')
  AND end_time <= sqlc.arg('end_time_to')
ORDER BY start_time;

-- name: UpdateTrip :exec
UPDATE trips
SET
  bus_id = sqlc.arg('bus_id'),
  route_id = sqlc.arg('route_id'),
  start_time = sqlc.arg('start_time'),
  end_time = sqlc.arg('end_time'),
  driven_distance_km = sqlc.arg('driven_distance_km'),
  energy_consumption_kWh = sqlc.arg('energy_consumption_kWh'),
  itcs_passengers_mean = sqlc.arg('itcs_passengers_mean'),
  itcs_passengers_min = sqlc.arg('itcs_passengers_min'),
  itcs_passengers_max = sqlc.arg('itcs_passengers_max'),
  grid_available_mean = sqlc.arg('grid_available_mean'),
  amb_temperature_mean = sqlc.arg('temperature_mean'),
  amb_temperature_min = sqlc.arg('temperature_min'),
  amb_temperature_max = sqlc.arg('temperature_max')
WHERE name = sqlc.arg('name');

-- name: DeleteTripByName :exec
DELETE FROM trips
WHERE name = sqlc.arg('name');

-- name: ListAllTrips :many
SELECT * FROM trips
ORDER BY start_time;

-- name: ListBuses :many
SELECT * FROM buses
ORDER BY bus_number;

-- name: ListRoutes :many
SELECT * FROM bus_routes
ORDER BY route_code;

-- name: GetBusRouteId :one
SELECT id FROM bus_routes WHERE route_code = sqlc.arg('bus_route');

-- name: GetBusRouteIdFromTripId :one
SELECT route_id FROM trips WHERE id = sqlc.arg('trip_id');

-- name: InsertTelemetry :copyfrom
INSERT INTO telemetry (
  trip_id,
  time,
  electric_power_demand,
  gnss_altitude,
  gnss_course,
  gnss_latitude,
  gnss_longitude,
  itcs_bus_route_id,
  itcs_number_of_passengers,
  itcs_stop_name,
  odometry_articulation_angle,
  odometry_steering_angle,
  odometry_vehicle_speed,
  odometry_wheel_speed_fl,
  odometry_wheel_speed_fr,
  odometry_wheel_speed_ml,
  odometry_wheel_speed_mr,
  odometry_wheel_speed_rl,
  odometry_wheel_speed_rr,
  status_door_is_open,
  status_grid_is_available,
  status_halt_brake_is_active,
  status_park_brake_is_active,
  temperature_ambient,
  traction_brake_pressure,
  traction_traction_force
)
VALUES (
  sqlc.arg('trip_id'),
  sqlc.arg('time'),
  sqlc.arg('electric_power_demand'),
  sqlc.arg('gnss_altitude'),
  sqlc.arg('gnss_course'),
  sqlc.arg('gnss_latitude'),
  sqlc.arg('gnss_longitude'),
  sqlc.arg('bus_route_id'),
  sqlc.arg('itcs_number_of_passengers'),
  sqlc.arg('itcs_stop_name'),
  sqlc.arg('odometry_articulation_angle'),
  sqlc.arg('odometry_steering_angle'),
  sqlc.arg('odometry_vehicle_speed'),
  sqlc.arg('odometry_wheel_speed_fl'),
  sqlc.arg('odometry_wheel_speed_fr'),
  sqlc.arg('odometry_wheel_speed_ml'),
  sqlc.arg('odometry_wheel_speed_mr'),
  sqlc.arg('odometry_wheel_speed_rl'),
  sqlc.arg('odometry_wheel_speed_rr'),
  sqlc.arg('status_door_is_open'),
  sqlc.arg('status_grid_is_available'),
  sqlc.arg('status_halt_brake_is_active'),
  sqlc.arg('status_park_brake_is_active'),
  sqlc.arg('temperature_ambient'),
  sqlc.arg('traction_brake_pressure'),
  sqlc.arg('traction_traction_force')
);

-- name: GetTelemetryByTrip :many
SELECT * FROM telemetry
WHERE trip_id = sqlc.arg('trip_id')
ORDER BY time;

-- name: ListTelemetryInRange :many
SELECT * FROM telemetry
WHERE trip_id = sqlc.arg('trip_id')
  AND time >= sqlc.arg('start_time')
  AND time <= sqlc.arg('end_time')
ORDER BY time;

-- name: DeleteTelemetryByTrip :exec
DELETE FROM telemetry
WHERE trip_id = sqlc.arg('trip_id');

-- name: MakePartitions :exec
CALL public.run_maintenance_proc();

