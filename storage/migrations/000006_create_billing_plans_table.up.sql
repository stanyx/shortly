CREATE TABLE public.billing_plans
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    name character varying COLLATE pg_catalog."default",
    description character varying COLLATE pg_catalog."default",
    price character varying COLLATE pg_catalog."default",
    CONSTRAINT billing_plans_pk PRIMARY KEY (id)
);

insert into public.billing_plans (name, description, price) values ('free', '', '0');
insert into public.billing_plans (name, description, price) values ('standart', '', '199');
insert into public.billing_plans (name, description, price) values ('pro', '', '2999');