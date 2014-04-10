CREATE SEQUENCE "public"."client_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 20
 CACHE 1;
ALTER TABLE "public"."client_seq" OWNER TO "mikasner";

CREATE SEQUENCE "public"."loginuser_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 20
 CACHE 1;
ALTER TABLE "public"."loginuser_seq" OWNER TO "mikasner";

CREATE SEQUENCE "public"."query_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 20
 CACHE 1;
ALTER TABLE "public"."query_seq" OWNER TO "mikasner";

CREATE SEQUENCE "public"."sale_seq"
 INCREMENT 1
 MINVALUE 1
 MAXVALUE 9223372036854775807
 START 20
 CACHE 1;
ALTER TABLE "public"."sale_seq" OWNER TO "mikasner";




CREATE TABLE "public"."client" (
"id" uuid NOT NULL,
"name" varchar(255) COLLATE "default",
"vatid" varchar COLLATE "default",
"address" varchar(255) COLLATE "default",
"zip" varchar(255) COLLATE "default",
"city" varchar(255) COLLATE "default",
"state" varchar(255) COLLATE "default",
"country" varchar(255) COLLATE "default",
"note" varchar(255) COLLATE "default",
"apikey" varchar(255) COLLATE "default",
"plan" varchar(255) COLLATE "default",
"querylimit" int4,
"taxreports" bool,
"trial" bool,
"trialend" timestamp(6),
"confirmed" bool,
"confirmkey" varchar(255) COLLATE "default",
"confirmkeyvalid" timestamp(6),
PRIMARY KEY ("id")
)
WITH (OIDS=FALSE);

ALTER TABLE "public"."client" OWNER TO "mikasner";


CREATE TABLE "public"."loginuser" (
"id" int8 DEFAULT nextval('loginuser_seq'::regclass) NOT NULL,
"name" varchar(255) COLLATE "default",
"email" varchar(255) COLLATE "default",
"password" varchar(255) COLLATE "default",
"active" boolean,
"client_id" uuid,
PRIMARY KEY ("id"),
FOREIGN KEY ("client_id") REFERENCES "public"."client" ("id") ON DELETE CASCADE ON UPDATE CASCADE
)
WITH (OIDS=FALSE);

ALTER TABLE "public"."loginuser" OWNER TO "mikasner";


CREATE TABLE "public"."sale" (
"id" int8 DEFAULT nextval('sale_seq'::regclass) NOT NULL,
"timestamp" timestamp(6),
"amount" int8,
"tax" int8,
"amountwithtax" int8,
"taxrate" int8,
"countrycode" varchar(255) COLLATE "default",
"client_id" uuid,
"createdby" int8,
"createdat" timestamp(6),
PRIMARY KEY ("id"),
FOREIGN KEY ("client_id") REFERENCES "public"."client" ("id") ON DELETE CASCADE ON UPDATE CASCADE,
FOREIGN KEY ("createdby") REFERENCES "public"."loginuser" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION
)
WITH (OIDS=FALSE);

ALTER TABLE "public"."sale" OWNER TO "mikasner";

CREATE TABLE "public"."query" (
"id" int8 DEFAULT nextval('query_seq'::regclass) NOT NULL,
"client_id" uuid,
"month" int2,
"year" int2,
"query" int8 NOT NULL,
PRIMARY KEY ("id"),
FOREIGN KEY ("client_id") REFERENCES "public"."client" ("id") ON DELETE CASCADE ON UPDATE CASCADE
)
WITH (OIDS=FALSE);

ALTER TABLE "public"."query" OWNER TO "mikasner";