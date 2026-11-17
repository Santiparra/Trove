# Seguridad y privacidad

Trove mantiene la biblioteca local. La versión actual no tiene cuentas propias, telemetría, integración de IA ni llamadas a proveedores de metadata. El frontend empaquetado no carga scripts o imágenes remotas arbitrarias.

SQLite y los respaldos no tienen cifrado propio: quien acceda a los archivos puede leerlos. La protección del dispositivo y de la cuenta de almacenamiento corresponde al entorno Windows y al servicio configurado. No se distribuyen credenciales, bibliotecas ni rutas personales dentro del ejecutable.

## Archivos y recuperación

- JSON y SQLite se identifican y validan antes de restaurar; versiones desconocidas, referencias rotas y contenido incompleto se rechazan.
- La restauración prepara una base aparte y requiere confirmación del perfil de destino.
- Se guarda una copia anterior y se intenta recuperarla si falla el reemplazo.
- Rutas de respaldo, logs, cachés y credenciales quedan fuera del modelo portable.
- El estado local de un respaldo no garantiza que un sincronizador de nube lo haya subido.

Importar futuras fuentes externas no debe ejecutar HTML o SQL aportado por ellas ni convertir metadata en identidad local. Las consultas a proveedores, cuando se incorporen, transmiten al menos la consulta/ID y exponen la IP al servicio; no equivalen a funcionamiento sin tráfico externo.

## Automatización

La validación de PR usa permisos de lectura. Solo el job de publicación sobre un push a `main` solicita escritura de contenido. Las Actions de terceros utilizadas están fijadas a revisiones concretas. No se publican releases desde PRs ni desde pushes a `dev`.

Las releases portables no están firmadas digitalmente. Los checksums sirven para comprobar integridad del paquete; no son cifrado ni sustituyen una firma del editor.
