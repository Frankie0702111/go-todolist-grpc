-- init.sql

-- Create a new database for unit tests
CREATE DATABASE test_db OWNER root;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE test_db TO root;