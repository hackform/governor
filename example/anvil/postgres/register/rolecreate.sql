CREATE ROLE "{{`{{name}}`}}" WITH LOGIN PASSWORD '{{`{{password}}`}}' VALID UNTIL '{{`{{expiration}}`}}';
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO "{{`{{name}}`}}";
