CREATE TABLE public.webhooks
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    account_id bigint NOT NULL,
    name character varying NOT NULL DEFAULT '',
    description character varying NOT NULL DEFAULT '',
    events character varying[] NOT NULL DEFAULT '{}',
    url character varying NOT NULL DEFAULT '',
    active boolean NOT NULL DEFAULT true,
    CONSTRAINT webhooks_pk PRIMARY KEY (id)
);