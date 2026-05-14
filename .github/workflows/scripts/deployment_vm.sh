#! /bin/bash

cd $PROJECT_PATH

echo "> Directory changed to: $(pwd)"

# Build new image and start new containers
sudo git pull
sudo docker compose -f docker-compose.prod.yml up -d --build --remove-orphans