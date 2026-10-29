create or replace function check_balance()
returns trigger as $$
declare
    total BIGINT;
begin
    select sum(amount) into total
    from ledger
    where ledger.transaction_id = new.transaction_id;

    if total != 0 then
        raise exception 'Transaction % does not balance.', new.transaction_id;
    end if;

    return new;
end;
$$ Language plpgsql;


create constraint trigger ensure_balance
after insert on ledger
deferrable initially deferred
for each row execute function check_balance();

-- By making it "Deferred", you are telling the database bouncer: 
-- "Hey, I'm about to do a bunch of inserts. Keep your eyes closed until I yell COMMIT. Then you can check my math."