-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE accounts DROP `xxx_unrecognized`, DROP `xxx_sizecache`;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE accounts ADD `xxx_unrecognized` varbinary(255) DEFAULT NULL, ADD `xxx_sizecache` int(11) DEFAULT NULL;
