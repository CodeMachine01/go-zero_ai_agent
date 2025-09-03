go run ./mcp/chat.go -f ./mcp/etc/chat.yaml

go run ./api/chat.go -f ./api/etc/chat.yaml

启动POSTgreSQL向量数据库

docker run -d --name my-postgres2 POSTGRES_USER=root -e POSTGRES_PASSWORD=123456 -e POSTGRES_DB=dayu_ai_agent -p 127.0.0.1:5432:5432 ankane/pgvector

本地redis服务启动

docker run -itd -p 6379:6379 --name redis -v /D/ProjectFiles/Docker/redis/redis.conf:/etc/redis/redis.conf -v /D/ProjectFiles/Docker/redis/data:/data redis redis-server /etc/redis/redis.conf