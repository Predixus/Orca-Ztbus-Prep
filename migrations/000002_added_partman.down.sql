-- Remove pg_partman configuration for telemetry table
DELETE FROM public.part_config WHERE parent_table = 'public.telemetry';
