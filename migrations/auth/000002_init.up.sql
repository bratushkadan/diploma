ALTER TABLE "auth"."users" ADD activated boolean NOT NULL DEFAULT false;

CREATE TABLE "auth"."confirmation_links" (
  email varchar(150) NOT NULL,
  content TEXT NOT NULL,
  expires_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (email) REFERENCES "auth"."users" (email) ON DELETE CASCADE
);
