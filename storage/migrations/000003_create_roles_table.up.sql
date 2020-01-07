CREATE TABLE public.roles
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    name character varying COLLATE pg_catalog."default",
    description character varying NOT NULL DEFAULT '' COLLATE pg_catalog."default",
    account_id bigint NOT NULL,
    CONSTRAINT roles_pk PRIMARY KEY (id)
);

ALTER TABLE public.roles
    ADD CONSTRAINT account_name_unique UNIQUE (name, account_id)
;