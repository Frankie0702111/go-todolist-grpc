-- init.sql

-- Create a new user for unit tests
CREATE USER test_user WITH PASSWORD 'test_password';

-- Create a new database for unit tests
CREATE DATABASE test_db OWNER test_user;

-- Grant privileges
GRANT ALL PRIVILEGES ON DATABASE test_db TO test_user;