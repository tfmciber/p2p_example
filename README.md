#Crear Modulo de go
go mod init tfm_ciber/alber/p2p
#Descargar todos los paquetes
go mod tidy
#Compilar el programa
go build


sysctl -w net.core.rmem_max=2500000


sudo apt-get install gcc-mingw-w64 -y
sudo apt-get install gcc-multilib -y

GOOS=windows GOARCH=386 \
  CGO_ENABLED=1 CXX=i686-w64-mingw32-g++ CC=i686-w64-mingw32-gcc \
  go build