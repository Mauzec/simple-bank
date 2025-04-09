BEGIN;

SELECT * FROM accounts WHERE id = 1;

INSERT INTO transfers (from_account_id, to_account_id, amount) VALUES (1, 2 ,10) RETURNING *;

INSERT INTO entries (account_id, amount) VALUES (1, -10) RETURNING *;
INSERT INTO entries (account_id, amount) VALUES (2, 10) RETURNING *;

SELECT * FROM accounts WHERE id = 1 FOR UPDATE;
UPDATE accounts SET balance = 0 WHERE id = 1 RETURNING *;

SELECT * FROM accounts WHERE id = 2 FOR UPDATE;
UPDATE accounts SET balance = 20 WHERE id = 2 RETURNING *;

ROLLBACK;