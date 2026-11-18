# Seguridad y privacidad

Trove mantiene la biblioteca local. No tiene cuentas propias, telemetría ni integración de IA. El catálogo opcional de Libros consulta Hardcover y complementa metadata con Open Library. El frontend empaquetado no carga scripts o imágenes remotas arbitrarias.

SQLite y los respaldos no tienen cifrado propio: quien acceda a los archivos puede leerlos. La protección del dispositivo y de la cuenta de almacenamiento corresponde al entorno Windows y al servicio configurado. No se distribuyen credenciales, bibliotecas ni rutas personales dentro del ejecutable.

## Archivos y recuperación

- JSON y SQLite se identifican y validan antes de restaurar; versiones desconocidas, referencias rotas y contenido incompleto se rechazan.
- La restauración prepara una base aparte y requiere confirmación del perfil de destino.
- Se guarda una copia anterior y se intenta recuperarla si falla el reemplazo.
- Rutas de respaldo, logs, cachés y credenciales quedan fuera del modelo portable.
- El estado local de un respaldo no garantiza que un sincronizador de nube lo haya subido.

## Catálogo opcional

Las búsquedas transmiten consulta o ISBN a Hardcover. Al seleccionar o refrescar una edición se consultan sus referencias y, cuando corresponde, el ISBN en Open Library. Las descargas de portadas exponen la IP al host de imágenes. No se consulta ni sincroniza la biblioteca personal de Hardcover.

Al iniciar se comprueba la API de Hardcover (solo si hay clave) con una consulta de una referencia de catálogo, y Open Library con un ISBN público fijo de diagnóstico. Estas comprobaciones también exponen la IP, pero no envían títulos, notas ni progreso de la biblioteca. Se ejecutan una vez en segundo plano; se pueden repetir manualmente y no hay sondeos periódicos de red. El estado, los errores genéricos y los tiempos se conservan solo en memoria. Las consultas normales alimentan el mismo diagnóstico; los mensajes no incorporan cuerpos remotos ni credenciales.

El modo solo local se carga antes de iniciar los chequeos y bloquea el transporte HTTP compartido por catálogos y portadas, incluyendo redirecciones. Activarlo cancela las solicitudes en curso, también durante la lectura del cuerpo de respuesta. Las lecturas de la base de datos y de portadas cacheadas no pasan por la red. Esta preferencia pertenece al perfil, queda fuera de respaldos y no elimina las credenciales.

La clave se introduce en un campo protegido y se guarda por perfil en `catalog-credentials.bin`, cifrada con DPAPI para la cuenta Windows. No se devuelve al frontend ni se incluye en respaldos. Guardarla no comprueba su validez hasta la consulta. Quitarla conserva la biblioteca y sus snapshots.

Go limita endpoints, HTTPS, redirecciones, tamaños y tiempos de espera; no devuelve cuerpos de error con credenciales. Las portadas se admiten solo desde `assets.hardcover.app`, `images.hardcover.app` y `covers.openlibrary.org`, se validan como imágenes raster y se sirven a React como data URL desde una caché local por perfil. No se envía la clave al descargar imágenes ni a Open Library. La metadata se presenta como texto, sin ejecutar HTML o SQL externo.

## Automatización

La validación de PR usa permisos de lectura. Solo el job de publicación sobre un push a `main` solicita escritura de contenido. Las Actions de terceros utilizadas están fijadas a revisiones concretas. No se publican releases desde PRs ni desde pushes a `dev`.

Las releases portables no están firmadas digitalmente. Los checksums sirven para comprobar integridad del paquete; no son cifrado ni sustituyen una firma del editor.
