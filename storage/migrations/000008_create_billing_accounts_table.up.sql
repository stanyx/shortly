CREATE TABLE public.billing_accounts
(
    account_id bigint,
    plan_id bigint,
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    active boolean NOT NULL DEFAULT false,
    CONSTRAINT billing_accounts_pk PRIMARY KEY (id)
);