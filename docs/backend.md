# Backend, persistencia y recuperación

SQLite está aceptado. «Backend» significa lógica local de la aplicación; no implica un servidor separado. El diseño físico y la política exacta de resguardo están pendientes.

## SQLite: qué implica

SQLite es una biblioteca embebida que lee y escribe una base almacenada en disco. No requiere instalar ni arrancar un servicio de base de datos ni gestionar puertos, usuarios o contraseñas de un servidor.

La app gestiona conexiones, migraciones y cierre. Según el modo de funcionamiento pueden existir archivos auxiliares de journal o WAL/SHM: copiar solo el archivo principal mientras está abierto no garantiza un respaldo completo.

El tamaño depende del contenido; no hay un servidor permanente asociado. Bibliotecas de cientos o miles de entradas son un uso razonable para SQLite. El consumo total de Trove dependerá también del contenedor, la interfaz y las integraciones; aún no hay mediciones.

## Experiencia confirmada

- Usar Trove, registrar cambios, cerrar y desentenderse: sin exportaciones manuales cotidianas.
- Experiencia portable preferida, con carpetas de salida de respaldos configurables.
- Archivos ligeros preparados por Trove para que un cliente de Drive/Dropbox los sincronice. El número exacto no es prioritario. No se pide integrar APIs de esos servicios.
- Respaldo de datos propios más exportación abierta de fallback, por ejemplo JSON.
- Datos propios y contenido liviano/textual en el resguardo. Portadas y otros recursos pesados permanecen locales, fuera del resguardo sincronizado, con referencias para recuperarlos. La metadata textual no se excluye por su procedencia externa; campos y límites exactos pendientes.
- No es indispensable editar datos con LibreOffice.

## Contenido del resguardo

Debe conservar los datos personales necesarios para reconstruir biblioteca y consumo: estados, puntajes, notas, progreso, fechas e historial dentro del alcance implementado; también la identidad y los vínculos necesarios para no perder su asociación con obras y ediciones.

El título local y las referencias externas forman parte del registro propio confirmado. Autor e identificadores de edición son candidatos a completar ese mínimo identificativo; los campos exactos están pendientes. Las copias de metadata textual pueden incluirse en el resguardo sin convertirse por ello en datos propios ni en su fuente de verdad.

Una referencia permite intentar recuperar una portada o metadata, pero no garantiza que sigan disponibles. Después de restaurar, los datos propios deben ser utilizables sin esperar esa recuperación; la metadata ausente puede volver a buscarse o vincularse manualmente.

## Mecanismo propuesto, pendiente de cerrar

1. Mantener datos activos fuera del área sincronizada. La separación entre datos personales y caché puede ser física o lógica; aún no se elige esquema.
2. Generar automáticamente un SQLite consistente con el contenido de resguardo y un fallback abierto restaurable, excluyendo recursos pesados según las reglas acordadas.
3. Preparar cada salida en un archivo temporal, validarlo y sustituir el anterior al terminar; asociar SQLite y fallback a una misma generación para detectar diferencias entre versiones sincronizadas.
4. Actualizar tras cambios con una breve espera para agrupar escrituras y completar lo pendiente al cierre normal. Un fallo de resguardo no debe borrar la última copia válida y debe ser visible para el usuario. Frecuencia, reintentos y política de versiones por acordar.
5. Restaurar desde esos archivos con comprobación de formato y compatibilidad, sin exigir conexión; recuperar recursos ausentes o reconstruir metadata posteriormente, por medio o proveedor.

Trove puede garantizar que prepara las salidas locales según el diseño acordado, pero la subida a la nube depende del cliente de sincronización y de la conectividad. El cierre forzado o apagado abrupto también debe contemplarse al concretar el intervalo de resguardo.

## Datos internos y exportación: decisiones independientes

El usuario aclaró que el número de archivos es secundario. La necesidad real es especializar los medios y poder reconstruir metadata por sección/proveedor sin afectar otras secciones.

Propuesta pendiente: una base SQLite activa con conceptos compartidos y tablas especializadas por medio, manteniendo claramente diferenciados datos propios y snapshots externos. Una base puede contener muchas tablas; no exige una estructura idéntica para libros, películas y series. Tablas concretas pendientes del modelado.

La independencia de reconstrucción se consigue delimitando las operaciones: seleccionar medio/proveedor, actualizar sus snapshots y recursos, conservar datos propios y no modificar otros medios. Si un proveedor atiende varios medios, se necesita también acotar por medio y entradas seleccionadas; filtrar solo por proveedor no basta.

La exportación puede ser global o por secciones independientemente de la organización interna. Reconstruir metadata por secciones no implica todavía aprobar restauración parcial de datos propios; esa funcionalidad y sus conflictos cruzados se definirían aparte.

## Responsabilidades futuras

Persistencia y migraciones; historial; importaciones; adaptadores externos; snapshots locales; resguardo automático y restauración. Bibliotecas, estructura de archivos, protección de credenciales y empaquetado portable dependen del stack y las decisiones pendientes.
