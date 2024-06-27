CREATE TABLE IF NOT EXISTS "public"."tasks" (
  "id" SERIAL PRIMARY KEY,
	"user_id" int4 NOT NULL,
	"category_id" int4 NOT NULL,
  "title" varchar(100) NOT NULL,
  "note" text,
  "url" text,
  "specify_datetime" timestamptz(6) DEFAULT NULL,
	"is_specify_time" bool DEFAULT FALSE,
	"priority" int2 DEFAULT 1 NOT NULL,
	"is_complete" bool DEFAULT FALSE,
  "created_at" timestamptz(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updated_at" timestamptz(6)
);

COMMENT ON COLUMN "public"."tasks"."title" IS '標題';
COMMENT ON COLUMN "public"."tasks"."note" IS '備註';
COMMENT ON COLUMN "public"."tasks"."url" IS '網址';
COMMENT ON COLUMN "public"."tasks"."specify_datetime" IS '指定日期時間(Y-m-d H:i:s)';
COMMENT ON COLUMN "public"."tasks"."is_specify_time" IS '是否指定時間 (0:不指定 1:指定)';
COMMENT ON COLUMN "public"."tasks"."priority" IS '優先度 (1:低 2:中 3:高)';
COMMENT ON COLUMN "public"."tasks"."is_complete" IS '是否完成';
COMMENT ON COLUMN "public"."tasks"."created_at" IS '新增時間';
COMMENT ON COLUMN "public"."tasks"."updated_at" IS '更新時間';

CREATE UNIQUE INDEX "title_uidx" ON "public"."tasks" USING btree (
  "title"
);

ALTER TABLE "public"."tasks" 
  ADD CONSTRAINT "users_user_id_foreign" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON DELETE CASCADE ON UPDATE NO ACTION,
  ADD CONSTRAINT "categories_category_id_foreign" FOREIGN KEY ("category_id") REFERENCES "public"."categories" ("id") ON DELETE CASCADE ON UPDATE NO ACTION;