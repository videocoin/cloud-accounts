-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE accounts DROP `balance`, DROP `balance_wei`;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE accounts ADD `balance` double DEFAULT NULL, ADD `balance_wei` varchar(255) DEFAULT NULL;
