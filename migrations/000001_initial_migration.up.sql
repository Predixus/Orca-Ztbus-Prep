
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
    energy_consumption_kWh REAL,
    
    -- ITCS passenger stats
    itcs_passengers_mean REAL,
    itcs_passengers_min INTEGER,
    itcs_passengers_max INTEGER,
    
    -- Grid status
    grid_available_mean REAL,

    -- Ambient temperature stats
    temperature_mean REAL,
    temperature_min REAL,
    temperature_max REAL
);

-- Indexes for query speed
CREATE INDEX idx_trips_start_time ON trips(start_time);
CREATE INDEX idx_trips_bus_id ON trips(bus_id);
CREATE INDEX idx_trips_route_id ON trips(route_id);
