CREATE TABLE IF NOT EXISTS "public"."users" (
  "id" SERIAL PRIMARY KEY,
  "username" varchar(32) NOT NULL,
  "email" varchar(64) NOT NULL,
  "password" varchar(255) NOT NULL,
  "status" bool DEFAULT TRUE,
  "created_at" timestamptz(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz(6)
);

COMMENT ON COLUMN "public"."users"."username" IS '用戶名';
COMMENT ON COLUMN "public"."users"."email" IS '信箱';
COMMENT ON COLUMN "public"."users"."password" IS '密碼';
COMMENT ON COLUMN "public"."users"."status" IS '狀態 (0:關閉 1:開啟)';
COMMENT ON COLUMN "public"."users"."created_at" IS '新增時間';
COMMENT ON COLUMN "public"."users"."updated_at" IS '更新時間';

CREATE UNIQUE INDEX "email_uidx" ON "public"."users" USING btree (
  "email"
);