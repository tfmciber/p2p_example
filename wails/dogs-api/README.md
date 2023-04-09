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

git ignore no va, hacer que si hay un error en el csv no se pare el bench
al hacer benchmark si no están el otro online da erro
limitar busqueda dht a 10 segundos --- hecho
port=5000
probability=0.02
sudo iptables -A INPUT -p udp --dport $port -m statistic --mode random --probability $probability -j DROP
sudo iptables -A INPUT -p tcp --dport $port -m statistic --mode random --probability $probability -j DROP
sudo iptables -F

wget https://go.dev/dl/go1.18.10.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.18.10.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
sudo apt install build-essential -y