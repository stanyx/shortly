CREATE TABLE public.billing_price
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    plan_id bigint NOT NULL,
    price character varying NOT NULL,
    is_annual boolean NOT NULL DEFAULT false,
    CONSTRAINT billing_price_pk PRIMARY KEY (id)
);

ALTER TABLE public.billing_price
    ADD CONSTRAINT billing_price_unique UNIQUE (plan_id, is_annual)
;