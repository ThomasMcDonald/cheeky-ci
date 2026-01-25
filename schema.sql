DROP TABLE IF EXISTS jobs;
CREATE TABLE jobs(
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL
)
