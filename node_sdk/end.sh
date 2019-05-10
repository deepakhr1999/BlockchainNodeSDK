# clean the keystore
rm -rf ./hfc-key-store

# launch network; create channel and join peer to channel
cd ../first-network
echo y | ./byfn.sh down
docker rm $(docker ps -a | awk '/dev/{print $1}')
docker rmi $(docker images | awk '/fabcar/{print $3}')