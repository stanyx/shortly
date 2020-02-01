ALTER TABLE public.billing_accounts ADD COLUMN charge integer;
ALTER TABLE public.billing_accounts ADD COLUMN is_annual boolean;
ALTER TABLE public.billing_plans ADD COLUMN stripe_id character varying;

CREATE TABLE public.stripe_products
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    stripe_id character varying NOT NULL,
    name character varying NOT NULL,
    created bigint NOT NULL,
    CONSTRAINT stripe_products_pk PRIMARY KEY (id)
);

CREATE TABLE public.stripe_plans
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    stripe_id character varying NOT NULL,
    name character varying NOT NULL,
    created bigint NOT NULL,
    CONSTRAINT stripe_plans_pk PRIMARY KEY (id)
);

CREATE TABLE public.stripe_subscriptions
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    stripe_id character varying NOT NULL,
    created bigint NOT NULL,
    account_id bigint NOT NULL,
    active boolean NOT NULL DEFAULT true,
    canceled_at bigint,
    CONSTRAINT stripe_payments_pk PRIMARY KEY (id)
);

CREATE TABLE public.stripe_charges
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    stripe_id character varying NOT NULL,
    created bigint NOT NULL,
    account_id bigint NOT NULL,
    CONSTRAINT stripe_charges_pk PRIMARY KEY (id)
);

CREATE TABLE public.stripe_customers
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    stripe_id character varying NOT NULL,
    created bigint NOT NULL,
    account_id bigint NOT NULL,
    CONSTRAINT stripe_customers_pk PRIMARY KEY (id)
);

CREATE TABLE public.stripe_events
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    stripe_id character varying NOT NULL,
    timestamp bigint NOT NULL,
    payload jsonb NOT NULL,
    error character varying DEFAULT '',
    CONSTRAINT stripe_events_pk PRIMARY KEY (id)
);