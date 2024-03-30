create table users (
    chat_id bigint primary key,
    username varchar(255),
    first_name varchar(255),
    last_name varchar(255),
    approved boolean default false,
    admin boolean default false,
    updated_at timestamp default current_timestamp,
    created_at timestamp default current_timestamp
);


-- median aggregate function
CREATE OR REPLACE FUNCTION _final_median(numeric[])
    RETURNS numeric AS
$$
SELECT AVG(val)
FROM (
         SELECT val
         FROM unnest($1) val
         ORDER BY 1
         LIMIT  2 - MOD(array_upper($1, 1), 2)
             OFFSET CEIL(array_upper($1, 1) / 2.0) - 1
     ) sub;
$$
    LANGUAGE 'sql' IMMUTABLE;

CREATE AGGREGATE median(numeric) (
    SFUNC=array_append,
    STYPE=numeric[],
    FINALFUNC=_final_median,
    INITCOND='{}'
);
