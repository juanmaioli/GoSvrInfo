# 📋 Plan de Implementación: GoSvrInfo

[Documentación oficial de Go](https://go.dev/doc/)

## 1. 🎯 Objetivo
Crear un binario en Go sin dependencias externas (100% biblioteca estándar) que recopile la información del sistema en Linux y la devuelva por consola en formato JSON idéntico a `host.json`.

---

## 2. 📝 Tareas Planificadas

1. **Inicialización del módulo Go:**
   - Crear `go.mod` con nombre de módulo `gosvrinfo`.
2. **Definición de estructuras de datos:**
   - Crear las estructuras con tags JSON exactos según `host.json`.
3. **Módulos de recolección de métricas:**
   - **Sistema y Tiempo:** `time`, `system_uptime` (`/proc/uptime`), `hostname` (`os.Hostname`).
   - **Geolocalización / Red Pública:** Consulta a `http://ip-api.com/json/` con timeout y fallback.
   - **CPU:** Información de `/proc/cpuinfo` y cálculo de uso mediante `/proc/stat`.
   - **RAM:** Información y cálculo de uso desde `/proc/meminfo`.
   - **Almacenamiento:** Particiones montadas y espacio vía `syscall.Statfs` / `/proc/mounts`.
   - **Interfaces de Red:** Nombre, dirección MAC e IP local vía `net.Interfaces()`.
   - **Sistema Operativo:** Parseo de `/etc/os-release` y kernel vía `syscall.Uname`.
   - **GPU:** Detección de NVIDIA (mediante comando `nvidia-smi` si existe) o búsqueda en `/sys/class/drm`.
4. **Pruebas y Verificación:**
   - Compilar con `go build`.
   - Ejecutar y validar que el JSON generado sea válido y cumpla con la estructura de `host.json`.
