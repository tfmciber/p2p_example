#Crear Modulo de go
go mod init tfm_ciber/alber/p2p
#Descargar todos los paquetes
go mod tidy
#Compilar el programa
go build

sysctl -w net.core.rmem_max=2500000
