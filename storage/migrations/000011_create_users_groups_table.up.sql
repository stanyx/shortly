CREATE TABLE public.users_groups
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    group_id bigint,
    user_id bigint,
    CONSTRAINT users_groups_pk PRIMARY KEY (id)
);