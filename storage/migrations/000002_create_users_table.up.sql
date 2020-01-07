CREATE TABLE public.users
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    username character varying NOT NULL COLLATE pg_catalog."default",
    password character varying NOT NULL COLLATE pg_catalog."default",
    phone character varying NOT NULL DEFAULT '' COLLATE pg_catalog."default",
    email character varying NOT NULL DEFAULT '' COLLATE pg_catalog."default",
    company character varying NOT NULL DEFAULT '' COLLATE pg_catalog."default",
    account_id bigint,
    is_staff boolean NOT NULL DEFAULT false,
    role_id bigint,
    CONSTRAINT users_username_unique UNIQUE (username)

);

ALTER TABLE public.users
    ADD CONSTRAINT users_pk PRIMARY KEY (id)
;