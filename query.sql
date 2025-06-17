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
  temperature_mean,
  temperature_min,
  temperature_max
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
  temperature_mean = sqlc.arg('temperature_mean'),
  temperature_min = sqlc.arg('temperature_min'),
  temperature_max = sqlc.arg('temperature_max')
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

