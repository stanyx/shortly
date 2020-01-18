insert into billing_accounts(account_id, plan_id, active, started_at, ended_at) 
select a.id, (select id from billing_plans where name = 'free' limit 1), 
true, now(), 
'2100-01-01T00:00:00'::timestamp with time zone
from accounts a
left join billing_accounts ba on ba.account_id = a.id and ba.active = true
where ba.id is null
order by ba.id