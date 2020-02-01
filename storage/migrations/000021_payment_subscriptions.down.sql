ALTER TABLE public.billing_accounts DROP COLUMN charge;
ALTER TABLE public.billing_accounts DROP COLUMN is_annual;

DROP TABLE public.stripe_products;
DROP TABLE public.stripe_plans;
DROP TABLE public.stripe_subscriptions;
DROP TABLE public.stripe_charges;