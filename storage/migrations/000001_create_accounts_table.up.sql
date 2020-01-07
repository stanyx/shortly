CREATE TABLE public.accounts
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    name character varying COLLATE pg_catalog."default" NOT NULL DEFAULT ''::character varying,
    confirmed boolean NOT NULL DEFAULT false,
    verified boolean NOT NULL DEFAULT false,
    created_at timestamp with time zone DEFAULT now()
);

ALTER TABLE public.accounts
    ADD CONSTRAINT accounts_pk PRIMARY KEY (id)
;