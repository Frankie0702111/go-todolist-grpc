CREATE TABLE IF NOT EXISTS "public"."verify_emails" (
  "id" SERIAL PRIMARY KEY,
  "user_id" int4 NOT NULL,
  "username" varchar(32) NOT NULL,
  "email" varchar(64) NOT NULL,
  "secret_code" varchar(255) NOT NULL,
  "is_used" bool DEFAULT FALSE,
  "expired_at" timestamptz NOT NULL DEFAULT (CURRENT_TIMESTAMP + interval '15 minutes'),
  "created_at" timestamptz(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz(6)
);

COMMENT ON COLUMN "public"."verify_emails"."username" IS '用戶名';
COMMENT ON COLUMN "public"."verify_emails"."email" IS '信箱';
COMMENT ON COLUMN "public"."verify_emails"."secret_code" IS '安全碼';
COMMENT ON COLUMN "public"."verify_emails"."is_used" IS '狀態 (0:未讀取 1:已開啟)';
COMMENT ON COLUMN "public"."verify_emails"."expired_at" IS '過期時間';
COMMENT ON COLUMN "public"."verify_emails"."created_at" IS '新增時間';
COMMENT ON COLUMN "public"."verify_emails"."updated_at" IS '更新時間';

ALTER TABLE "public"."verify_emails" ADD CONSTRAINT "users_user_id_foreign_verify" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE "public"."users" ADD COLUMN "is_email_verified" bool DEFAULT FALSE;
COMMENT ON COLUMN "public"."users"."is_email_verified" IS '狀態 (0:未驗證 1:已驗證)';