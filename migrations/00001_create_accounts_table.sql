-- +goose Up
-- SQL in this section is executed when the migration is applied.
CREATE TABLE `accounts` (
  `id` varchar(36) NOT NULL,
  `user_id` varchar(255) DEFAULT NULL,
  `address` varchar(42) DEFAULT NULL,
  `key` varchar(512) DEFAULT NULL,
  `balance` double DEFAULT NULL,
  `updated_at` timestamp NULL DEFAULT NULL,
  `balance_wei` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
DROP TABLE accounts;