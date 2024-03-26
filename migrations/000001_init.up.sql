create table cars (
    id serial primary key,
    manufacturer text not null,
    model text not null,
    year integer not null,
    mileage integer not null,
    engine float not null,
    fuel text not null,
    drive text not null,
    automatic boolean not null,
    power integer not null,
    color text not null,
    price integer not null,
    description text not null,
    ad_id text not null,
    address text not null default '',
    link text not null,
    posted timestamp not null,
    parsed date not null default current_date
);


-- unique index for ad_id and date
create unique index cars_ad_id_parsed_idx on cars (ad_id, parsed);
