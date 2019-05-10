echo "Killing and removing containers:"
docker rm -f $(docker ps -aq)
echo "Removing volumes"
docker volume rm $(docker volume ls | awk '/net_/{print $2}')
DOCKER_IMAGE_IDS=$(docker images | awk '($1 ~ /dev-peer.*.mycc.*/) {print $3}')
if [ -z "$DOCKER_IMAGE_IDS" -o "$DOCKER_IMAGE_IDS" == " " ]; then
	echo "---- No images available for deletion ----"
else
	docker rmi -f $DOCKER_IMAGE_IDS
fi
echo y | docker network prune
echo "Removing unwanted folders: channel-artifacts/*.block channel-artifacts/*.tx crypto-config"
rm -rf channel-artifacts/*.block channel-artifacts/*.tx crypto-config
rm -f docker-compose-e2e.yaml
