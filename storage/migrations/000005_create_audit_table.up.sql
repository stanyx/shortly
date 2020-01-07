CREATE TABLE public.audit
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    entity character varying NOT NULL COLLATE pg_catalog."default",
    entity_id bigint NOT NULL,
    action character varying NOT NULL COLLATE pg_catalog."default",
    "timestamp" timestamp with time zone,
    snapshot jsonb,
    CONSTRAINT audit_pk PRIMARY KEY (id)
);