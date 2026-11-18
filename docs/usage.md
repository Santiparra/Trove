# Uso y datos locales

## Bibliotecas de Libros y Novelas ligeras

Cada sección abre su portal con estados, lecturas actuales, últimas finalizaciones conocidas y distribución visual. Comparten recorridos, pero sus listados, búsquedas y series son independientes. Las vistas previas se limitan al espacio disponible; los listados completos siguen accesibles.

Cada volumen de novela ligera se registra individualmente y puede vincularse a una serie y número de orden. Una serie con el mismo nombre en Libros y Novelas ligeras no se fusiona automáticamente. Cambiar de pestaña no reclasifica ni modifica registros existentes.

- La búsqueda muestra primero resultados locales por título, autoría, serie o identificador. En Libros agrega candidatos de Hardcover en el mismo flotante cuando hay clave configurada, sin bloquear las coincidencias locales. Enter o «Ver todos los resultados» abre el listado completo.
- Elegir un estado abre sus registros. Los botones superiores cambian el filtro.
- La serie aparece sobre el título. Una serie local conserva sus libros ordenados por posición (los que no tienen posición quedan al final) y, en Libros, suma candidatos del catálogo por nombre. Una serie de catálogo abre la búsqueda compartida por su nombre.
- Pulsar un autor en una ficha local o remota abre la búsqueda compartida de sus libros. Los resultados remotos requieren la clave configurada; las coincidencias locales siguen disponibles sin ella.
- La ficha permite guardar notas y modificar favorito, propiedad y estado sin abrir el formulario completo.
- Solo el título es obligatorio al crear un libro. ISBN, autoría y fechas pueden completarse después.
- En Novelas ligeras se usa «Añadir volumen», con la misma edición mínima y datos personales. La ficha permite editar, marcar favorito/propiedad, guardar notas y cambiar estados, incluida A continuación.
- Eliminar exige confirmación y elimina los datos asociados al libro, no otros libros de su serie.

## Catálogo opcional de Libros

1. Configurar una clave de lectura de Hardcover desde **Claves**, junto al indicador de respaldos. Guardarla no consulta ni sincroniza la biblioteca personal del proveedor.
2. Usar la búsqueda principal por título, autor, serie o ISBN. Elegir un resultado abre directamente la ficha compartida, con portada, información y gestión. Una obra abre preferentemente la edición local vinculada; si no existe, se elige la edición con más campos informados entre las devueltas por el proveedor. Buscar un ISBN concreto respeta esa edición.
3. Desde la ficha, usar **Elegir otra edición** o **Cambiar edición vinculada** para ver hasta 20 ediciones y buscar otra por ISBN. Consultar y revisar no agrega registros.
4. El botón **Agregar · Por leer** agrega con un clic. Elegir otro estado en su desplegable ejecuta el alta inmediatamente, sin una confirmación adicional; los registros existentes también cambian de estado al elegirlo. Una única serie bibliográfica se preselecciona; se puede quitar, y varias asociaciones requieren elección. Agregar como Leído no fabrica fechas de consumo.
5. Una edición ya vinculada abre su registro local con estado y acciones, sin ofrecer agregar. Coincidencias por ISBN o por otra edición de la obra se revisan antes de vincular. No se fusionan automáticamente ediciones distintas.

La ficha distribuye portada, información y gestión en tres columnas cuando hay espacio, sin deformar la imagen. La descripción extensa se expande con **Leer más**. Los identificadores y fuentes están en un bloque desplegable al pie.

La búsqueda flotante espera 140 ms de pausa antes de consultar localmente y otros 350 ms después de obtener esos resultados antes de consultar Hardcover. Seguir escribiendo cancela los temporizadores pendientes; las respuestas antiguas no reemplazan la consulta vigente. Las peticiones que ya salieron no se cancelan en la red.

Desde la ficha se puede vincular o cambiar edición, refrescar metadata y desvincular. Estas operaciones conservan los datos personales; el refresco fallido conserva el snapshot anterior. La descripción y demás metadata guardada se pueden leer offline. Las portadas ya descargadas se reutilizan localmente; las no disponibles muestran un marcador. Open Library solo complementa campos faltantes con un ISBN seguro de la edición elegida.

El alta manual sigue disponible sin clave ni conexión. Novelas ligeras conserva su flujo manual. La clave y las imágenes cacheadas no se exportan: tras restaurar en otra ubicación puede ser necesario configurar la clave y descargar portadas nuevamente.

## Salud de proveedores

### Modo solo local

Desde el panel **Proveedores**, el botón **Modo solo local** desactiva las consultas a Hardcover y Open Library, los chequeos de salud y las descargas de portadas. El bloqueo se aplica en Go, también cancela solicitudes en curso y se guarda por perfil para respetarlo al reiniciar. No borra claves ni lo interpreta como una caída del proveedor.

La biblioteca, los snapshots persistidos, las portadas ya cacheadas y toda la gestión personal siguen disponibles. Las portadas no descargadas muestran su marcador; refrescar metadata y buscar ediciones remotas queda deshabilitado. Al reactivar la red se vuelven a permitir consultas, conservando las reglas del circuit breaker.

Para comprobar una restauración sin conexión: activar el modo, restaurar la generación elegida y recorrer las fichas locales. La configuración de red no forma parte del respaldo y no cambia al restaurarlo.

### Indicador y diagnóstico

El indicador **Proveedores**, junto a Claves y respaldos, muestra el último estado conocido de las APIs de Hardcover y Open Library. Al abrirlo se ven disponibilidad, duración de la última consulta, hora y posibles pausas. Una clave ausente o rechazada se distingue de un fallo de servicio. No comprueba las portadas ni garantiza que todas las operaciones del proveedor funcionen.

- Al iniciar se ejecuta una comprobación breve por API en segundo plano, con presupuesto de 8 segundos y sin bloquear la biblioteca. Hardcover se omite si falta clave. Las consultas normales actualizan el mismo estado.
- **Comprobar ahora** permite repetir el diagnóstico, como máximo una vez cada 15 segundos; respeta las pausas de cada proveedor. No hay sondeos remotos periódicos. El indicador refresca únicamente el estado local cada 3 segundos mientras la ventana es visible.
- Tres fallos consecutivos observados pausan nuevas consultas durante un minuto. Después se admite una única consulta de recuperación; si funciona se reabre el acceso, y si falla se vuelve a pausar. Las solicitudes todavía en cola también respetan el circuito.
- Un HTTP 429 pausa de inmediato, con al menos un minuto y respetando `Retry-After` hasta 24 horas. Un rechazo de acceso de Hardcover bloquea nuevas consultas hasta guardar nuevamente o reemplazar la clave. Quitarla muestra «Sin configurar».
- Un registro ausente (404) no marca el proveedor como caído. Las cancelaciones locales no suman fallos. La duración excluye la espera local impuesta por el límite de ritmo.

El diagnóstico no modifica la biblioteca ni sus snapshots; el estado de salud es temporal y no se exporta. Un problema de Open Library no lo convierte en proveedor principal ni impide usar los datos ya guardados.

## Estados y fechas

Estados: Por leer, A continuación, Leyendo, Leído, Pausado y Sin terminar. A continuación se asigna manualmente.

Al entrar en Leyendo se registra inicio si falta. Al entrar en Leído desde otro estado se registra finalización si hay inicio y falta finalización. Pausar y retomar conserva las fechas existentes. Marcar un libro antiguo como leído sin inicio no genera una finalización ficticia.

Las fechas son editables. La finalización no puede preceder al inicio; pueden coincidir el mismo día. Guardar explícitamente una fecha o borrarla prevalece sobre los automatismos de esa edición. Restaurar no aplica esos automatismos.

Los últimos leídos requieren fecha de finalización para establecer un orden cronológico. Los registros sin ella permanecen en el listado de leídos.

## Perfiles

| Perfil | Ubicación predeterminada |
| --- | --- |
| `habitual` | `data/habitual/` junto al ejecutable de usuario |
| `test` | `data/test/` junto al ejecutable de prueba |
| `development` | `.local/development/` bajo el directorio de ejecución |

El indicador superior muestra perfil, ruta, versión y revisión. También se puede consultar la identidad sin abrir datos:

```powershell
.\Trove.exe --version
```

Una raíz alternativa siempre recibe la subcarpeta del perfil:

```powershell
.\Trove.exe --profile test --data-dir "C:\TrovePruebas"
```

El ejemplo usa `C:\TrovePruebas\test\`. No compartir el mismo directorio activo entre instancias durante una restauración.

## Avisos de operaciones

Los éxitos cotidianos se muestran como avisos flotantes sin desplazar el contenido. Se cierran a los 5 segundos, con pausa mientras tienen el foco o el puntero, y pueden descartarse manualmente. Los errores permanecen en el contexto de la operación.

## Respaldo automático

1. Abrir **Respaldo y restauración**.
2. Elegir una carpeta distinta de los datos activos. Se agrega el nombre del perfil.
3. Se genera una primera copia. Las modificaciones posteriores se agrupan durante dos segundos y se respaldan automáticamente.

Se conservan dos generaciones completas, cada una con `trove.sqlite`, `trove.json` y `complete.json`. Si no cambió el contenido no se publica una generación duplicada. El cierre normal espera los guardados aceptados y completa pendientes. Un cierre forzado puede dejar cambios solo en la base activa; al reiniciar se comprueba su contenido.

Un fallo de salida conserva el cambio local y la última copia válida. Se informa el error y se reintenta. «Al día» confirma los archivos locales, no el estado de un sincronizador externo. Los respaldos no incluyen portadas, cachés, logs, credenciales ni rutas de configuración.

## Restaurar

1. Abrir **Respaldo y restauración**. El selector muestra las generaciones completas válidas de la carpeta configurada como **Actual** y **Anterior**, con fecha y hora locales.
2. Elegir la generación y pulsar **Revisar respaldo seleccionado**. La preparación vuelve a validar esa generación exacta; si cambió o ya no está disponible, se informa el problema sin elegir otra automáticamente.
3. Revisar fecha, contenido y perfil de destino, marcar la confirmación y pulsar **Restaurar y reemplazar biblioteca**. **Cambiar selección** descarta la preparación sin reemplazar datos.

También se puede usar **Seleccionar respaldo para restaurar** para elegir un SQLite o JSON generado por Trove desde el diálogo nativo. Cualquiera de las representaciones válidas funciona sin la otra ni el manifiesto al elegirla manualmente. No es una importación desde servicios externos.

La preparación se valida antes del reemplazo y vence a los diez minutos. Se conserva una copia `before-restore-*.sqlite` en el perfil. Si falla el reemplazo se intenta reabrir la biblioteca anterior. La configuración de salida permanece ligada al perfil de destino.

Si una base dañada impide iniciar, abrir una raíz nueva y restaurar allí sin modificar la anterior:

```powershell
.\Trove.exe --profile test --data-dir "C:\TroveRecuperacion"
```

Los formatos compatibles, límites y relaciones están en [formato de respaldo](backup-format.md).

El esquema local actual es 4. La migración deja una copia `before-schema-4-*.sqlite` antes de incorporar identificadores normalizados y snapshots bibliográficos; conserva el identificador original, identidades y fechas desconocidas. Los respaldos nuevos usan formato 3 y contienen ambos medios y la metadata vinculada. Se pueden recuperar formatos 1 y 2; el formato 1 se interpreta como Libros. Restaurar siempre reemplaza el perfil completo, sin mezclar entradas posteriores.
