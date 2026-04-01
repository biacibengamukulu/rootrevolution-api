---
Deployment Complete

Access routes:
- All routes public: GET /products, GET /products/:id, GET /products/authorize/:token
- Auth required: POST, PUT, DELETE on products

Live at: https://cloudcalls.easipath.com/backend_rootrevolution/api

Docker setup on server: /apps/docker-compose-script/rootrevolution/

Port 3351 was confirmed free and is now in use.

Status at startup:
- Cassandra connected, all 4 tables migrated
- 33 products seeded from mockup/product_data.json
- Default admin user created: biangacila@gmail.com / Nathan010309*

Quick reference:
# On server
cd /apps/docker-compose-script/rootrevolution
docker compose logs -f          # follow logs
docker compose restart          # restart
docker compose down && docker compose up -d  # full restart

# Local redeploy after code changes
rsync -avz --exclude='.git' --exclude='.idea' \
/home/biangacila/GolandProjects/rootrevolution-api/ \
safer:/apps/docker-compose-script/rootrevolution/
ssh safer "cd /apps/docker-compose-script/rootrevolution && docker compose up -d --build"