CREATE TABLE tasks (
  id        serial primary key,
  user_id   int NOT NULL,
  task      varchar(255) NOT NULL,
  complete  int DEFAULT(0)
)
