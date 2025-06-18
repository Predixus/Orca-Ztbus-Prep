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
    driven_distance_km REAL,
    energy_consumption_kWh INTEGER,
    
    -- ITCS passenger stats
    itcs_passengers_mean REAL,
    itcs_passengers_min INTEGER,
    itcs_passengers_max INTEGER,
    
    -- Grid status
    grid_available_mean REAL,

    -- Ambient temperature stats
    amb_temperature_mean REAL,
    amb_temperature_min REAL,
    amb_temperature_max REAL
);

-- Trip telemetry
CREATE TABLE telemetry (
    id SERIAL PRIMARY KEY,
    trip_id INTEGER NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    time TIMESTAMP NOT NULL,

    electric_power_demand REAL,
    temperature_ambient REAL,
    traction_brake_pressure REAL,
    traction_traction_force REAL,
    
    -- GNSS
    gnss_altitude REAL,
    gnss_course REAL,
    gnss_latitude REAL,
    gnss_longitude REAL,

    -- itcs
    itcs_bus_route_id INTEGER REFERENCES bus_routes(id) ON DELETE CASCADE,
    itcs_number_of_passengers INTEGER,
    itcs_stop_name TEXT,

    -- Odemetry
    odometry_articulation_angle REAL,
    odometry_steering_angle REAL,
    odometry_vehicle_speed REAL,
    odometry_wheel_speed_fl REAL,
    odometry_wheel_speed_fr REAL,
    odometry_wheel_speed_ml REAL,
    odometry_wheel_speed_mr REAL,
    odometry_wheel_speed_rl REAL,
    odometry_wheel_speed_rr REAL,
  
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
