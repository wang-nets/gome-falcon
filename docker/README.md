## Running gome-falcon container

    docker pull openfalcon/gome-falcon:0.2.0
    docker run -itd -p 8081:8081 openfalcon/gome-falcon:0.2.0 bash /run.sh hbs

## Running gome-falcon container with docker-compose

    docker-compose -f init.yml up -d gome-falcon

## Running mysql and redis container

    docker-compose -f init.yml up -d mysql redis

## Stop and Remove containers

    docker-compose -f init.yml rm -f
