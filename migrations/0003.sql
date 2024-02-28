CREATE TABLE IF NOT EXISTS "categories" (
    "id" SERIAL PRIMARY KEY,
    "name" VARCHAR(255) UNIQUE NOT NULL
);
CREATE TABLE IF NOT EXISTS "categorization" (
    "category_id" INT,
    "question_id" INT,
    CONSTRAINT "fk_category_id" FOREIGN KEY ("category_id") REFERENCES "categories"("id"),
    CONSTRAINT "fk_question_id" FOREIGN KEY ("question_id") REFERENCES "questions"("id"),
    PRIMARY KEY("category_id", "question_id")
);
CREATE TYPE difficulty AS ENUM ('easy', 'medium', 'hard');
ALTER TABLE "questions"
ADD "difficulty" difficulty;
UPDATE "questions"
SET "difficulty" = 'medium';
ALTER TABLE "questions"
ALTER COLUMN "difficulty"
SET NOT NULL;