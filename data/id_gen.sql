-- unique ID generator in PostgreSQL
CREATE SEQUENCE public.table_id_seq START 5000;

CREATE OR REPLACE FUNCTION public.next_t_id(OUT result bigint) AS $$
DECLARE
    our_epoch bigint := 1314220021721;
    seq_id bigint;
    now_millis bigint;
    shard_id int := 5;
BEGIN
    SELECT nextval('public.table_id_seq') % 1024 INTO seq_id;
    SELECT FLOOR(EXTRACT(EPOCH FROM clock_timestamp()) * 1000) INTO now_millis;
    result := (now_millis - our_epoch) << 23;
    result := result | (shard_id <<10);
    result := result | (seq_id);
END;
    $$ LANGUAGE PLPGSQL;

select public.next_t_id(), public.next_t_id(), nextval('public.table_id_seq');