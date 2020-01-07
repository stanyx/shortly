CREATE TABLE public.tags
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    tag character varying COLLATE pg_catalog."default",
    link_id bigint,
    CONSTRAINT tags_pk PRIMARY KEY (id)
);