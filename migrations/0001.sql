CREATE TABLE IF NOT EXISTS "questions" (
    "id" SERIAL PRIMARY KEY,
    "text" VARCHAR(255) UNIQUE NOT NULL
);
CREATE TABLE IF NOT EXISTS "answers" (
    "id" SERIAL PRIMARY KEY,
    "text" VARCHAR(255) NOT NULL,
    "is_correct" BOOLEAN NOT NULL,
    "question_id" INT,
    CONSTRAINT "fk_question_id" FOREIGN KEY ("question_id") REFERENCES "questions"("id"),
    UNIQUE ("text", "question_id")
);