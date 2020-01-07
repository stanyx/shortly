CREATE TABLE public.billing_options
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    name character varying COLLATE pg_catalog."default",
    description character varying COLLATE pg_catalog."default",
    value character varying COLLATE pg_catalog."default",
    plan_id bigint,
    CONSTRAINT billing_options_pk PRIMARY KEY (id)
);

insert into public.billing_options (name, description, value, plan_id) values ('url_limit', '', '1000', 1);
insert into public.billing_options (name, description, value, plan_id) values ('timedata_limit', '', '1', 1);
insert into public.billing_options (name, description, value, plan_id) values ('users_limit', '', '1', 1);

insert into public.billing_options (name, description, value, plan_id) values ('url_limit', '', '10000', 2);
insert into public.billing_options (name, description, value, plan_id) values ('timedata_limit', '', '60', 2);
insert into public.billing_options (name, description, value, plan_id) values ('users_limit', '', '10', 2);
insert into public.billing_options (name, description, value, plan_id) values ('tags', '', '1', 2);
insert into public.billing_options (name, description, value, plan_id) values ('tags_limit', '', '500', 2);
insert into public.billing_options (name, description, value, plan_id) values ('groups', '', '1', 2);
insert into public.billing_options (name, description, value, plan_id) values ('groups_limit', '', '500', 2);
insert into public.billing_options (name, description, value, plan_id) values ('campaigns', '', '1', 2);
insert into public.billing_options (name, description, value, plan_id) values ('campaigns_limit', '', '100', 2);

insert into public.billing_options (name, description, value, plan_id) values ('url_limit', '', '500000', 3);
insert into public.billing_options (name, description, value, plan_id) values ('timedata_limit', '', '365', 3);
insert into public.billing_options (name, description, value, plan_id) values ('users_limit', '', '100', 3);
insert into public.billing_options (name, description, value, plan_id) values ('tags', '', '1', 3);
insert into public.billing_options (name, description, value, plan_id) values ('tags_limit', '', '10000', 3);
insert into public.billing_options (name, description, value, plan_id) values ('groups', '', '1', 3);
insert into public.billing_options (name, description, value, plan_id) values ('groups_limit', '', '10000', 3);
insert into public.billing_options (name, description, value, plan_id) values ('campaigns', '', '1', 3);
insert into public.billing_options (name, description, value, plan_id) values ('campaigns_limit', '', '500', 3);