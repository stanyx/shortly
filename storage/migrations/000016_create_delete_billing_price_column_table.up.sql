ALTER TABLE public.billing_plans DROP COLUMN price;

insert into billing_price (plan_id, price, is_annual) values (1, '0', false);
insert into billing_price (plan_id, price, is_annual) values (1, '0', true);
insert into billing_price (plan_id, price, is_annual) values (2, '290', false);
insert into billing_price (plan_id, price, is_annual) values (2, '2400', true);
insert into billing_price (plan_id, price, is_annual) values (3, '2990', false);
insert into billing_price (plan_id, price, is_annual) values (3, '24000', true);