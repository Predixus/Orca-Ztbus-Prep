
-- Drop indexes
DROP INDEX IF EXISTS idx_trips_route_id;
DROP INDEX IF EXISTS idx_trips_bus_id;
DROP INDEX IF EXISTS idx_trips_start_time;

-- Drop tables
DROP TABLE IF EXISTS trips;
DROP TABLE IF EXISTS buses;
DROP TABLE IF EXISTS bus_routes;

