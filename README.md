# 🖥️ GoSvrInfo

[Documentación oficial de Go](https://go.dev/doc/)

Herramienta de línea de comandos en **Go** para Linux que recopila información completa del hardware, sistema operativo, estado de recursos, interfaces de red y geolocalización pública, devolviéndola en formato JSON estructurado.

Desarrollada **100% con la biblioteca estándar de Go**, sin librerías externas ni dependencias de runtime.

---

## 1. ⚙️ Características

* **Rendimiento y portabilidad:** Binario estático ultra ligero.
* **Información del Sistema:** Uptime, fecha en timestamp Unix y hostname.
* **Geolocalización e IP Pública:** Consulta segura y con timeout a `ip-api.com`.
* **CPU:** Modelo, núcleos/hilos, frecuencia en GHz y porcentaje de uso en tiempo real.
* **Memoria RAM:** Capacidad total y porcentaje de memoria en uso.
* **Almacenamiento:** Particiones y discos físicos montados con tamaño total y espacio libre.
* **Red:** Interfaces activas con dirección MAC e IP local asignada.
* **Sistema Operativo:** Distribución Linux, versión y release del kernel.
* **GPU:** Detección de aceleradoras NVIDIA (`nvidia-smi`) o adaptadores gráficos vía PCI/DRM.

---

## 2. 🚀 Compilación e Instalación

### Compilación local
```bash
go build -o gosvrinfo main.go
```

### Compilación estática (portátil para cualquier Linux)
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gosvrinfo main.go
```

### Instalación en el sistema
```bash
sudo install -m 755 gosvrinfo /usr/local/bin/gosvrinfo
```

---

## 3. 📋 Ejemplo de Salida

Al ejecutar `gosvrinfo`, genera una salida JSON estándar por pantalla:

```json
{
  "time": 1791888288,
  "system_uptime": 1683260,
  "hostname": "AramisWks",
  "ip_public_address": "181.27.250.202",
  "city": "Neuquén",
  "country": "Argentina",
  "latitude": -38.9491,
  "longitude": -68.0627,
  "timezone": "America/Argentina/Salta",
  "isp": "Telefonica de Argentina",
  "CPU": "Intel(R) Core(TM) i5-10400F CPU @ 2.90GHz",
  "cores": 12,
  "cpu_speed": "4.02 GHz",
  "cpu_usage": "3%",
  "RAM": "31 GB",
  "ram_usage": "36%",
  "disk_space": [
    {
      "name": "nvme0n1p2",
      "size": "215 GB",
      "free": "148 GB"
    }
  ],
  "network_interfaces": [
    {
      "name": "enp3s0",
      "mac_address": "b4:2e:99:f4:15:f3",
      "ip_address": "172.21.5.5"
    }
  ],
  "operating_system": {
    "name": "Debian GNU/Linux",
    "version": "13 (trixie)",
    "kernel_version": "6.12.100+deb13-amd64"
  },
  "GPU": {
    "name": "NVIDIA GeForce GTX 1050",
    "driver_version": "550.163.01",
    "memory": "2 GB"
  }
}
```
