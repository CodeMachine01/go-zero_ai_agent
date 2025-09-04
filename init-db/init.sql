-- 启用必要的扩展
CREATE EXTENSION IF NOT EXISTS vector;
CREATE EXTENSION IF NOT EXISTS cube;
CREATE EXTENSION IF NOT EXISTS earthdistance;
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- 创建向量存储表
CREATE TABLE IF NOT EXISTS "public"."vector_store" (
    "id" BIGSERIAL PRIMARY KEY,
    "chat_id" varchar(255) NOT NULL,
    "role" varchar(50) NOT NULL,
    "content" TEXT NOT NULL,
    "embedding" JSONB NOT NULL,
    "source_type" VARCHAR(50) NOT NULL DEFAULT 'message',
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT now()
    );

-- 创建知识库表
CREATE TABLE IF NOT EXISTS "public"."knowledge_base" (
    "id" BIGSERIAL PRIMARY KEY,
    "title" VARCHAR(255) NOT NULL,
    "content" TEXT NOT NULL,
    "embedding" JSONB NOT NULL,
    "created_at" TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_vector_store_chat_id ON vector_store (chat_id);
CREATE INDEX IF NOT EXISTS idx_vector_store_created_at ON vector_store (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_knowledge_base_title ON knowledge_base (title);