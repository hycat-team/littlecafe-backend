-- Development bootstrap only. Production credentials are provisioned outside Git.
CREATE ROLE littlecafe_commerce LOGIN PASSWORD '12345678';
CREATE ROLE littlecafe_experience LOGIN PASSWORD '12345678';
CREATE ROLE littlecafe_ai LOGIN PASSWORD '12345678';
CREATE ROLE littlecafe_platform LOGIN PASSWORD '12345678';

CREATE DATABASE littlecafe_commerce OWNER littlecafe_commerce;
CREATE DATABASE littlecafe_experience OWNER littlecafe_experience;
CREATE DATABASE littlecafe_ai OWNER littlecafe_ai;
CREATE DATABASE littlecafe_platform OWNER littlecafe_platform;

\connect littlecafe_commerce
CREATE EXTENSION IF NOT EXISTS pgcrypto;
\connect littlecafe_experience
CREATE EXTENSION IF NOT EXISTS pgcrypto;
\connect littlecafe_ai
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS vector;
\connect littlecafe_platform
CREATE EXTENSION IF NOT EXISTS pgcrypto;
