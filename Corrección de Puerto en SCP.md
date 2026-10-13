 Corrección de Puerto en SCP 

  El flag para especificar el puerto en scp es -P (mayúscula) y debe ir antes de la ruta de destino (el
  -p en minúscula preserva fechas/permisos y causaría un error de sintaxis).
  ──────
  ### 1. 🚀 Comando Correcto

    scp -P 10019 /home/juan/Documentos/Dev/Apps/GoSvrInfo/gosvrinfo administrator@93.127.142.5:/tmp/
  ──────
  ### 2. 📲 Moverlo a /usr/local/bin en el Servidor

  Una vez transferido a /tmp/:

    ssh -p 10019 -t administrator@93.127.142.5 "sudo install -m 755 /tmp/gosvrinfo
  /usr/local/bin/gosvrinfo && rm /tmp/gosvrinfo"