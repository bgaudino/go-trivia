CREATE TABLE IF NOT EXISTS "users" (
    "id" SERIAL PRIMARY KEY,
    "username" VARCHAR(255) UNIQUE NOT NULL,
    "password" VARCHAR(255) NOT NULL
);
CREATE TABLE IF NOT EXISTS "sessions" (
    "id" SERIAL PRIMARY KEY,
    "token" VARCHAR(255) NOT NULL,
    "expiry" TIMESTAMP NOT NULL,
    "user_id" INT NOT NULL,
    CONSTRAINT "fk_user_id" FOREIGN KEY ("user_id") REFERENCES "users"("id")
);
