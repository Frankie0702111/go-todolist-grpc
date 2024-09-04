ALTER TABLE "public"."verify_emails" DROP CONSTRAINT "users_user_id_foreign_verify";

DROP TABLE IF EXISTS "public"."verify_emails";

ALTER TABLE "public"."users" DROP COLUMN "is_email_verified";