#! /bin/bash

cd $PROJECT_PATH

echo "> Directory changed to: $(pwd)"

# Build new image and start new containers
# Fetch latest changes and reset to the specific commit
sudo git fetch origin
sudo git reset --hard $COMMIT_SHA
sudo docker compose -f docker-compose.prod.yml up -d --build --remove-orphans