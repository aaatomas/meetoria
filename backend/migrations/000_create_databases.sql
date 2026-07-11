-- Additional databases for Keycloak and notification workers
CREATE DATABASE keycloak;
CREATE DATABASE meetoria_sms;
CREATE DATABASE meetoria_email;

-- Grant access to the meetoria user
GRANT ALL PRIVILEGES ON DATABASE keycloak TO meetoria;
GRANT ALL PRIVILEGES ON DATABASE meetoria_sms TO meetoria;
GRANT ALL PRIVILEGES ON DATABASE meetoria_email TO meetoria;
