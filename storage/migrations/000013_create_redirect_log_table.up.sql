CREATE TABLE public.redirect_log
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    short_url character varying COLLATE pg_catalog."default",
    long_url character varying COLLATE pg_catalog."default",
    headers jsonb,
    "timestamp" timestamp with time zone,
    CONSTRAINT redirect_log_pk PRIMARY KEY (id)
);