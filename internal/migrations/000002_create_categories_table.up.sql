CREATE TABLE IF NOT EXISTS "public"."categories" (
  "id" SERIAL PRIMARY KEY,
  "name" varchar(128) NOT NULL,
  "created_at" timestamptz(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz(6)
);

COMMENT ON COLUMN "public"."categories"."name" IS '類別名稱';
COMMENT ON COLUMN "public"."categories"."created_at" IS '新增時間';
COMMENT ON COLUMN "public"."categories"."updated_at" IS '更新時間';

CREATE UNIQUE INDEX "name_uidx" ON "public"."categories" USING btree (
  "name"
);