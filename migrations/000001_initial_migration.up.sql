-- Routes
CREATE TABLE bus_routes (
    id SERIAL PRIMARY KEY,
    route_code TEXT UNIQUE
);

-- Busses
CREATE TABLE buses (
    id SERIAL PRIMARY KEY,
    bus_number TEXT UNIQUE
);

-- Trips
CREATE TABLE trips (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    bus_id INTEGER REFERENCES buses(id) ON DELETE CASCADE,
    route_id INTEGER REFERENCES bus_routes(id) ON DELETE SET NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    driven_distance_km NUMERIC,
    energy_consumption_kWh NUMERIC,
    
    -- ITCS passenger stats
    itcs_passengers_mean NUMERIC,
    itcs_passengers_min INTEGER,
    itcs_passengers_max INTEGER,
    
    -- Grid status
    grid_available_mean NUMERIC,

    -- Ambient temperature stats
    temperature_mean NUMERIC,
    temperature_min NUMERIC,
    temperature_max NUMERIC
);

-- Trip telemetry
CREATE TABLE telemetry (
    id SERIAL PRIMARY KEY,
    trip_id INTEGER NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    time TIMESTAMP NOT NULL,

    electric_power_demand NUMERIC,
    temperature_ambient NUMERIC,
    traction_brake_pressure NUMERIC,
    traction_traction_force NUMERIC,
    
    -- GNSS
    gnss_altitude NUMERIC,
    gnss_course NUMERIC,
    gnss_latitude NUMERIC,
    gnss_longitude NUMERIC,

    -- itcs
    itcs_bus_route TEXT,
    itcs_number_of_passengers NUMERIC,
    itcs_stop_name TEXT,

    -- Odemetry
    odometry_articulation_angle NUMERIC,
    odometry_steering_angle NUMERIC,
    odometry_vehicle_speed NUMERIC,
    odometry_wheel_speed_fl NUMERIC,
    odometry_wheel_speed_fr NUMERIC,
    odometry_wheel_speed_ml NUMERIC,
    odometry_wheel_speed_mr NUMERIC,
    odometry_wheel_speed_rl NUMERIC,
    odometry_wheel_speed_rr NUMERIC,
  
    -- Statuses
    status_door_is_open BOOLEAN,
    status_grid_is_available BOOLEAN,
    status_halt_brake_is_active BOOLEAN,
    status_park_brake_is_active BOOLEAN
);

-- Indexes for query speed
CREATE INDEX idx_trips_start_time ON trips(start_time);
CREATE INDEX idx_trips_bus_id ON trips(bus_id);
CREATE INDEX idx_trips_route_id ON trips(route_id);
CREATE INDEX idx_telemetry_trip_time ON telemetry(trip_id, time);
