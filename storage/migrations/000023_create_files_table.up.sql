CREATE TABLE public.files
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    name character varying NOT NULL,
    content bytea,
    downloaded_at timestamp with time zone,
    CONSTRAINT files_pk PRIMARY KEY (id)
);