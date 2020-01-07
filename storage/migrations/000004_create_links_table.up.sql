CREATE TABLE public.links
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    short_url character varying(20) COLLATE pg_catalog."default" NOT NULL,
    long_url text COLLATE pg_catalog."default" NOT NULL,
    account_id bigint DEFAULT 0,
    description character varying COLLATE pg_catalog."default",
    CONSTRAINT urls_pkey PRIMARY KEY (id)
);

ALTER TABLE public.links
    ADD CONSTRAINT short_url_unique UNIQUE (short_url)
;