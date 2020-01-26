CREATE TABLE public.dashboards
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    account_id bigint NOT NULL,
    name character varying NOT NULL DEFAULT '',
    description character varying NOT NULL DEFAULT '',
    width integer NOT NULL,
    height integer NOT NULL,
    CONSTRAINT dashboards_pk PRIMARY KEY (id)
);

CREATE TABLE public.dashboards_widgets
(
    id bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1 ),
    account_id bigint NOT NULL,
    dashboard_id bigint NOT NULL,
    widget_type character varying NOT NULL DEFAULT '',
    title character varying NOT NULL DEFAULT '',
    data_url character varying NOT NULL DEFAULT '',
    x integer NOT NULL,
    y integer NOT NULL,
    span integer DEFAULT 1,
    CONSTRAINT dashboards_widgets_pk PRIMARY KEY (id)
);