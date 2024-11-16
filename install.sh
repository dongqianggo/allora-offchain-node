echo "allora-offchain-node installed!"

# Update
sudo apt update && sudo apt upgrade -y

# Install the build tools needed to run a node (curl, git, jq, lz4, and build-essential)
sudo apt -qy install curl git jq lz4 build-essential screen -y

# Install Docker via the following command
sudo apt install docker.io -y

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.29.2/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose

# We also need to assign the Docker Compose directory higher permissions.
sudo chmod +x /usr/local/bin/docker-compose

# Install Docker Compose CLI Plugin
DOCKER_CONFIG=${DOCKER_CONFIG:-$HOME/.docker}
mkdir -p $DOCKER_CONFIG/cli-plugins
curl -SL https://github.com/docker/compose/releases/download/v2.20.2/docker-compose-linux-x86_64 -o $DOCKER_CONFIG/cli-plugins/docker-compose

# Make plugin executable
chmod +x $DOCKER_CONFIG/cli-plugins/docker-compose

# Verify installation
docker --version
docker-compose --version


git clone https://github.com/allora-network/coin-prediction-reputer
cd coin-prediction-reputer

cp config.example.json config.json

chmod +x init.config
./init.config

docker compose up -d--build


git clone https://github.com/allora-network/basic-coin-prediction-node
cd basic-coin-prediction-node

cp .env.example .env

chmod +x init.config
./init.config