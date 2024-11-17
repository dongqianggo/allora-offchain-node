echo "------------------------allora installed begin!------------------------"
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


echo " installed coin-prediction-reputer begin!>>>>>>>>>>>>>>>>>>>>>>>>>>>"
git clone https://github.com/dongqianggo/coin-prediction-reputer
cd coin-prediction-reputer
cp config.example.json config.json
chmod +x init.config
./init.config
docker compose up -d
cd ../
echo " installed coin-prediction-reputer end!>>>>>>>>>>>>>>>>>>>>>>>>>>>"

echo " installed basic-coin-prediction-node begin!>>>>>>>>>>>>>>>>>>>>>>>>>>>"
git clone https://github.com/dongqianggo/basic-coin-prediction-node.git
cd basic-coin-prediction-node
chmod +x init.config
./init.config
docker compose up -d
cd ../
echo " installed basic-coin-prediction-node end!>>>>>>>>>>>>>>>>>>>>>>>>>>>"

echo " installed allora-offchain-node begin!>>>>>>>>>>>>>>>>>>>>>>>>>>>"
git clone https://github.com/dongqianggo/allora-offchain-node.git
cd allora-offchain-node.git
chmod +x init.config
./init.config
docker compose up -d
cd ../
echo " installed allora-offchain-node end!>>>>>>>>>>>>>>>>>>>>>>>>>>>"

echo "------------------------allora installed end!------------------------"