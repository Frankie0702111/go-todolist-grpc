ALTER TABLE "public"."tasks" 
  DROP CONSTRAINT IF EXISTS "users_user_id_foreign",
  DROP CONSTRAINT IF EXISTS "categories_category_id_foreign";

DROP INDEX IF EXISTS "title_uidx";
DROP TABLE IF EXISTS "public"."tasks";