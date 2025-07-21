-- Remove pg_partman configuration for telemetry table
DELETE FROM public.part_config WHERE parent_table = 'public.telemetry';

-- Drop all telemetry partitions
SELECT public.drop_partition_table(
    p_parent_table => 'public.telemetry',
    p_keep_table => false
);
