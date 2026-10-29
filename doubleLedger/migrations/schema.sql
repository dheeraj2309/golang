-- create account table

CREATE TABLE accounts(
    id UUID PRIMARY KEY,
    name varchar(255) not null,
    cached_balance BIGINT not null default 0,
    created_at timestamp default CURRENT_TIMESTAMP
);

-- create transaction table
create table transactions(
    id UUID PRIMARY KEY,
    reference varchar(255) unique,
    created_at timestamp default CURRENT_TIMESTAMP
);

-- create the ledger
create table ledger(
    id UUID PRIMARY KEY,
    transaction_id UUID REFERENCES transactions(id),
    account_id UUID REFERENCES accounts(id),
    amount BIGINT not null,
    created_at timestamp default CURRENT_TIMESTAMP
);
