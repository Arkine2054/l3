CREATE OR REPLACE FUNCTION log_item_update() RETURNS trigger AS $$
DECLARE
    uid_text text := current_setting('app.current_user_id', true);
    uid int;
BEGIN
    IF uid_text IS NULL OR uid_text = '' THEN
        uid := 1;
    ELSE
        uid := uid_text::int;
    END IF;

    INSERT INTO history(item_id, user_id, action, old_data, new_data, timestamp)
    VALUES (
               OLD.id,
               uid,
               'update',
               row_to_json(OLD),
               row_to_json(NEW),
               CURRENT_TIMESTAMP
           );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_items_update ON items;
CREATE TRIGGER trg_items_update
    AFTER UPDATE ON items
    FOR EACH ROW
EXECUTE FUNCTION log_item_update();


CREATE OR REPLACE FUNCTION log_item_delete() RETURNS trigger AS $$
DECLARE
    uid_text text := current_setting('app.current_user_id', true);
    uid int;
BEGIN
    IF uid_text IS NULL OR uid_text = '' THEN
        uid := 1;
    ELSE
        uid := uid_text::int;
    END IF;

    INSERT INTO history(item_id, user_id, action, old_data, new_data, timestamp)
    VALUES (
               OLD.id,
               uid,
               'delete',
               row_to_json(OLD),
               NULL,
               CURRENT_TIMESTAMP
           );
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_items_delete ON items;
CREATE TRIGGER trg_items_delete
    AFTER DELETE ON items
    FOR EACH ROW
EXECUTE FUNCTION log_item_delete();
