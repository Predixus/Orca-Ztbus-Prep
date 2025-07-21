-- Set up pg_partman for daily partitioning
SELECT public.create_parent(
    p_parent_table => 'public.telemetry',
    p_control => 'time',
    p_type => 'range',
    p_interval => '1 month',
    p_premake => 1,  
    p_start_partition => '2019-01-01'  -- Starting from beginning of dataset
);

SELECT public.run_maintenance();
