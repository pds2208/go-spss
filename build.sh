
docker build -t lfs-spss:v1 .
# Remove the intermediate image
docker image prune -f --filter label=stage=builder