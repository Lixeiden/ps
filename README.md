# ps
Go/Gin pastebin app 

* docker build -t ps_image .
* docker volume create ps_volume
* docker run -p 8001:8080 --name=ps_container --volume=ps_volume:/app/pst ps_image -d
