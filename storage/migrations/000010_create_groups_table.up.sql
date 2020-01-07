CREATE TABLE public.groups
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    name character varying COLLATE pg_catalog."default",
    description text COLLATE pg_catalog."default" DEFAULT ''::text,
    account_id bigint,
    CONSTRAINT groups_pk PRIMARY KEY (id)
);

CREATE TABLE public.links_groups
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    group_id bigint,
    link_id bigint,
    CONSTRAINT links_groups_pk PRIMARY KEY (id)
);