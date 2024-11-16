echo "------------------------allora installed begin!------------------------"


echo " installed coin-prediction-reputer begin!>>>>>>>>>>>>>>>>>>>>>>>>>>>"
git clone https://github.com/allora-network/coin-prediction-reputer
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
cp .env.example .env
cp config.example.json config.json
chmod +x init.config
./init.config
docker compose up -d
cd ../
echo " installed basic-coin-prediction-node end!>>>>>>>>>>>>>>>>>>>>>>>>>>>"

echo " installed allora-offchain-node begin!>>>>>>>>>>>>>>>>>>>>>>>>>>>"
git clone https://github.com/dongqianggo/allora-offchain-node.git
cd allora-offchain-node.git
cp .env.example .env
cp config.example.json config.json
chmod +x init.config
./init.config
docker compose up -d
cd ../
echo " installed allora-offchain-node end!>>>>>>>>>>>>>>>>>>>>>>>>>>>"


echo "------------------------allora installed end!------------------------"