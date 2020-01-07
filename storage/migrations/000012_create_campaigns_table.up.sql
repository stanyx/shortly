CREATE TABLE public.campaigns
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    name character varying NOT NULL COLLATE pg_catalog."default",
    description character varying COLLATE pg_catalog."default",
    active boolean NOT NULL DEFAULT true,
    account_id bigint NOT NULL,
    CONSTRAINT campaigns_pk PRIMARY KEY (id)
);

CREATE TABLE public.campaigns_links
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    link_id bigint,
    campaign_id bigint,
    utm_source character varying COLLATE pg_catalog."default",
    utm_medium character varying COLLATE pg_catalog."default",
    utm_term character varying COLLATE pg_catalog."default",
    utm_content character varying COLLATE pg_catalog."default",
    CONSTRAINT campaigns_links_pk PRIMARY KEY (id)
)