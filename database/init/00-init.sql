-- Development bootstrap only. Production credentials are provisioned outside Git.
CREATE ROLE littlecafe_identity LOGIN PASSWORD '12345678';
CREATE ROLE littlecafe_order LOGIN PASSWORD '12345678';
CREATE ROLE littlecafe_activity LOGIN PASSWORD '12345678';
CREATE ROLE littlecafe_wallet LOGIN PASSWORD '12345678';
CREATE ROLE littlecafe_ai LOGIN PASSWORD '12345678';
CREATE ROLE littlecafe_notification LOGIN PASSWORD '12345678';
CREATE ROLE littlecafe_reporting LOGIN PASSWORD '12345678';

CREATE DATABASE littlecafe_identity OWNER littlecafe_identity;
CREATE DATABASE littlecafe_order OWNER littlecafe_order;
CREATE DATABASE littlecafe_activity OWNER littlecafe_activity;
CREATE DATABASE littlecafe_wallet OWNER littlecafe_wallet;
CREATE DATABASE littlecafe_ai OWNER littlecafe_ai;
CREATE DATABASE littlecafe_notification OWNER littlecafe_notification;
CREATE DATABASE littlecafe_reporting OWNER littlecafe_reporting;

\\connect littlecafe_identity
CREATE EXTENSION IF NOT EXISTS pgcrypto;
\\connect littlecafe_order
CREATE EXTENSION IF NOT EXISTS pgcrypto;
\\connect littlecafe_activity
CREATE EXTENSION IF NOT EXISTS pgcrypto;
\\connect littlecafe_wallet
CREATE EXTENSION IF NOT EXISTS pgcrypto;
\\connect littlecafe_ai
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS vector;
\\connect littlecafe_notification
CREATE EXTENSION IF NOT EXISTS pgcrypto;
\\connect littlecafe_reporting
CREATE EXTENSION IF NOT EXISTS pgcrypto;
