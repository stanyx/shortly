ALTER TABLE public.redirect_log ADD COLUMN ip_addr character varying;
ALTER TABLE public.redirect_log ADD COLUMN country character varying;
ALTER TABLE public.redirect_log ADD COLUMN referer character varying;