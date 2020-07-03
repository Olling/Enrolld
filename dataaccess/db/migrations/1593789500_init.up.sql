CREATE TABLE servers (
    ID VARCHAR(64) PRIMARY KEY,
    IP VARCHAR(15),
    LastSeen TIMESTAMP,
    Groups TEXT[],
    Properties JSON
)