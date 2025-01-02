-- Create the database
CREATE DATABASE ecomm0001;

-- Grant all privileges on the database to the user 'postgres'
GRANT ALL PRIVILEGES ON DATABASE ecomm0001 TO postgres;

-- Set the timezone for the database to your local timezone
ALTER DATABASE ecomm0001 SET timezone TO 'Asia/Kolkata';

