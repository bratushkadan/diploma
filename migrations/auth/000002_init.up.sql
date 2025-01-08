ALTER TABLE "auth"."users" ADD activated boolean NOT NULL DEFAULT false;

CREATE TABLE "auth"."confirmation_ids" (
  email varchar(150) NOT NULL,
  id TEXT UNIQUE NOT NULL,
  expires_at timestamp NOT NULL,
  FOREIGN KEY (email) REFERENCES "auth"."users" (email) ON DELETE CASCADE
);
