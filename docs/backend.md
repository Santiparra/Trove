# Backend, persistencia y recuperación

SQLite está aceptado. «Backend» significa lógica local de la aplicación; no implica un servidor separado. El diseño físico y la política exacta de resguardo están pendientes.

## SQLite: qué implica

SQLite es una biblioteca embebida que lee y escribe una base almacenada en disco. No requiere instalar ni arrancar un servicio de base de datos ni gestionar puertos, usuarios o contraseñas de un servidor.

La app gestiona conexiones, migraciones y cierre. Según el modo de funcionamiento pueden existir archivos auxiliares de journal o WAL/SHM: copiar solo el archivo principal mientras está abierto no garantiza un respaldo completo.

El tamaño depende del contenido; no hay un servidor permanente asociado. Bibliotecas de cientos o miles de entradas son un uso razonable para SQLite. El consumo total de Trove dependerá también del contenedor, la interfaz y las integraciones; aún no hay mediciones.

## Experiencia confirmada

- Usar Trove, registrar cambios, cerrar y desentenderse: sin exportaciones manuales cotidianas.
- Experiencia portable preferida, con carpetas de salida de respaldos configurables.
- Pocos archivos ligeros preparados por Trove para que un cliente de Drive/Dropbox los sincronice. No se pide integrar APIs de esos servicios.
- Respaldo de datos propios más exportación abierta de fallback, por ejemplo JSON.
- Metadata externa y portadas fuera del resguardo sincronizado. Pueden permanecer locales y recuperarse después mediante referencias persistidas.
- No es indispensable editar datos con LibreOffice.

## Contenido del resguardo

Debe conservar los datos personales necesarios para reconstruir biblioteca y consumo: estados, puntajes, notas, progreso, fechas e historial dentro del alcance implementado; también la identidad y los vínculos necesarios para no perder su asociación con obras y ediciones.

Las referencias a proveedores y de portadas son información de recuperación, aunque la metadata descriptiva esté excluida. Propuesta pendiente: conservar además un mínimo identificativo local —título, autor e identificadores de edición disponibles— para reconocer entradas y buscar alternativas si el proveedor desaparece. No confundir este mínimo con una copia completa de metadata.

Una referencia permite intentar recuperar una portada o metadata, pero no garantiza que sigan disponibles. Después de restaurar, los datos propios deben ser utilizables sin esperar esa recuperación; la metadata ausente puede volver a buscarse o vincularse manualmente.

## Mecanismo propuesto, pendiente de cerrar

1. Mantener datos activos fuera del área sincronizada. La separación entre datos personales y caché puede ser física o lógica; aún no se elige esquema.
2. Generar automáticamente un SQLite consistente solo con el contenido de resguardo y un fallback abierto restaurable, sin copiar indiscriminadamente la base activa si incluye metadata excluida.
3. Preparar cada salida en un archivo temporal, validarlo y sustituir el anterior al terminar; asociar SQLite y fallback a una misma generación para detectar diferencias entre versiones sincronizadas.
4. Actualizar tras cambios con una breve espera para agrupar escrituras y completar lo pendiente al cierre normal. Un fallo de resguardo no debe borrar la última copia válida y debe ser visible para el usuario. Frecuencia, reintentos y política de versiones por acordar.
5. Restaurar desde esos archivos con comprobación de formato y compatibilidad, sin exigir conexión; recuperar metadata e imágenes posteriormente.

Trove puede garantizar que prepara las salidas locales según el diseño acordado, pero la subida a la nube depende del cliente de sincronización y de la conectividad. El cierre forzado o apagado abrupto también debe contemplarse al concretar el intervalo de resguardo.

## Número de archivos: decisión abierta

El usuario sugirió como ejemplo un SQLite y un JSON por tipo de medio. No implica todavía decidir una base activa por medio.

Alternativas: un par global para todos los medios, o un par por medio. La primera reduce archivos y facilita relaciones entre medios; la segunda permite respaldar/restaurar secciones pero exige resolver dependencias compartidas y consistencia entre archivos. Ambas pueden excluir portadas y metadata. Consultar antes de elegir.

## Responsabilidades futuras

Persistencia y migraciones; historial; importaciones; adaptadores externos; snapshots locales; resguardo automático y restauración. Bibliotecas, estructura de archivos, protección de credenciales y empaquetado portable dependen del stack y las decisiones pendientes.
