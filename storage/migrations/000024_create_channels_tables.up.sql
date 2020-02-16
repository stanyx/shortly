CREATE TABLE public.channels
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    account_id bigint NOT NULL,
    name character varying NOT NULL,
    CONSTRAINT channels_pk PRIMARY KEY (id)
);

CREATE TABLE public.campaigns_channels
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    campaign_id bigint NOT NULL,
    channel_id bigint NOT NULL,
    CONSTRAINT campaigns_channels_pk PRIMARY KEY (id)
);

ALTER TABLE public.campaigns_links RENAME COLUMN campaign_id TO chan_campaign_id;
ALTER TABLE public.campaigns_links RENAME TO campaigns_channels_links;