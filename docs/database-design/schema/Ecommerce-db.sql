CREATE TABLE "Products" (
  "id" uuid PRIMARY KEY,
  "name" varchar,
  "description" varchar,
  "origin" varchar,
  "color" enum,
  "size" enum,
  "shop_id" uuid,
  "inventory_id" uuid UNIQUE,
  "origin_price" bigint,
  "sale_price" bigint
);

CREATE TABLE "Category" (
  "id" uuid PRIMARY KEY,
  "name" varchar,
  "image_url" varchar,
  "description" varchar
);

CREATE TABLE "Comment" (
  "id" uuid PRIMARY KEY,
  "created_by" char,
  "content" varchar,
  "images" jsonb,
  "rated" bigint,
  "product_id" uuid
);

CREATE TABLE "Voucher" (
  "id" uuid PRIMARY KEY,
  "name" varchar,
  "description" varchar,
  "start_date" datetime,
  "expire_on" datetime,
  "amount" bigint
);

CREATE TABLE "notification" (
  "id" uuid PRIMARY KEY,
  "user_id" uuid,
  "title" varchar,
  "content" varchar
);

CREATE TABLE "shopping_carts" (
  "id" uuid PRIMARY KEY,
  "product_id" uuid,
  "amount" bigint,
  "user_id" uuid
);

CREATE TABLE "order" (
  "id" uuid PRIMARY KEY,
  "user_id" uuid,
  "status" bigint,
  "created_at" timestamp
);

CREATE TABLE "inventory" (
  "id" uuid PRIMARY KEY,
  "remain_product_number" bigint,
  "status" enum
);

CREATE TABLE "Shop_info" (
  "id" uuid PRIMARY KEY,
  "name" varchar,
  "address" varchar,
  "phone_number" varchar,
  "follower_number" bigint,
  "joined_at" timestamp,
  "rated" bigint,
  "user_id" uuid UNIQUE
);

CREATE TABLE "category_product" (
  "category_id" uuid,
  "product_id" uuid,
  PRIMARY KEY ("category_id", "product_id")
);

CREATE TABLE "order_items" (
  "order_id" uuid,
  "product_id" uuid,
  "amount" bigint,
  PRIMARY KEY ("order_id", "product_id")
);

CREATE TABLE "product_voucher" (
  "product_id" uuid,
  "voucher_id" uuid,
  PRIMARY KEY ("product_id", "voucher_id")
);

CREATE TABLE "user" (
  "id" uuid PRIMARY KEY,
  "user_name" varchar,
  "password" varchar,
  "role" enum,
  "mail" varchar,
  "avatar_url" binary,
  "first_name" varchar,
  "last_name" varchar,
  "dob" date,
  "phone_number" varchar,
  "male" enum,
  "setting_id" uuid
);

CREATE TABLE "profile" (
  "id" uuid PRIMARY KEY,
  "user_id" uuid UNIQUE,
  "avatar_url" varchar,
  "first_name" varchar,
  "last_name" varchar,
  "dob" varchar,
  "phone_number" varchar,
  "male" enum
);

CREATE TABLE "Address" (
  "id" uuid PRIMARY KEY,
  "street_no" bigint,
  "street" varchar,
  "ward" varchar,
  "district" varchar,
  "city" varchar,
  "user_id" uuid,
  "type" enum
);

CREATE TABLE "Setting" (
  "id" uuid PRIMARY KEY,
  "settings" jsonb,
  "user_id" uuid UNIQUE
);

CREATE TABLE "Payment" (
  "id" uuid PRIMARY KEY,
  "order_id" uuid UNIQUE,
  "status" enum,
  "paid_at" timestamp,
  "user_id" uuid UNIQUE,
  "type" enum,
  "amount" bigint,
  "transaction_id" uuid UNIQUE,
  "currency" enum
);

ALTER TABLE "order_items" ADD FOREIGN KEY ("product_id") REFERENCES "Products" ("id");

ALTER TABLE "Products" ADD FOREIGN KEY ("shop_id") REFERENCES "Shop_info" ("id");

ALTER TABLE "notification" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "category_product" ADD FOREIGN KEY ("category_id") REFERENCES "Category" ("id");

ALTER TABLE "order_items" ADD FOREIGN KEY ("order_id") REFERENCES "order" ("id");

ALTER TABLE "order" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "shopping_carts" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "category_product" ADD FOREIGN KEY ("product_id") REFERENCES "Products" ("id");

ALTER TABLE "Comment" ADD FOREIGN KEY ("product_id") REFERENCES "Products" ("id");

ALTER TABLE "Products" ADD FOREIGN KEY ("inventory_id") REFERENCES "inventory" ("id");

ALTER TABLE "shopping_carts" ADD FOREIGN KEY ("product_id") REFERENCES "Products" ("id");

ALTER TABLE "product_voucher" ADD FOREIGN KEY ("voucher_id") REFERENCES "Voucher" ("id");

ALTER TABLE "product_voucher" ADD FOREIGN KEY ("product_id") REFERENCES "Products" ("id");

ALTER TABLE "Address" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "Setting" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "Shop_info" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "profile" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "Payment" ADD FOREIGN KEY ("order_id") REFERENCES "order" ("id");

ALTER TABLE "Payment" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");
