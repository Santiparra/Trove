# Formato de respaldo de Trove

La versión **3 del formato de respaldo** es independiente del esquema SQLite (actualmente 4) y de la versión pública de la aplicación. Se mantiene lectura y restauración de las copias de formato 1 y 2.

## Organización de archivos

```text
Carpeta elegida/
  test/                       # o development / habitual
    generation-<id>/
      trove.sqlite
      trove.json
      complete.json
    generation-<id-anterior>/
      trove.sqlite
      trove.json
      complete.json
```

Se conservan las dos últimas generaciones completas y válidas. `complete.json` contiene versión de publicación, generación, fecha UTC, hash del contenido lógico y hashes SHA-256 de ambos archivos. Permite detectar pares incompletos o alterados antes de tratarlos como generaciones válidas para retención. No es una firma ni cifrado.

La generación se publica tras terminar los archivos. Una sincronización externa puede descargarlos por separado: **un JSON o SQLite válido de Trove funciona de forma independiente**, sin requerir el otro archivo ni el manifiesto. No confundir una copia incompleta del sincronizador con una generación terminada.

## JSON lógico, versión 3

Codificación UTF-8. Raíz:

| Campo | Tipo / significado |
| --- | --- |
| `format` | Cadena constante `trove-library` |
| `version` | Entero `3` en nuevas generaciones |
| `generation` | Identificador textual de la generación; no es el ID de una obra |
| `createdAt` | Fecha/hora de generación RFC 3339 en UTC |
| `catalog` | Contenido lógico descrito debajo |

`catalog` contiene cuatro arrays. Se generan arrays vacíos para colecciones sin registros. Los IDs son cadenas, no posiciones en un array ni secuencias autoincrementales dependientes de un motor.

| Colección | Campos |
| --- | --- |
| `series` | `id`, `name`, `medium` |
| `works` | `id`, `medium`, `title`, `authors` (array de cadenas), `notes`, `addedOn`, `favorite`; `seriesId` y `position` opcionales |
| `editions` | `id`, `workId`, `language`, `format`, `identifier`, `owned`; `identifiers` y `catalog` opcionales |
| `cycles` | `id`, `editionId`, `status`, `startedOn`, `finishedOn` |

En esta versión, `medium` admite `books` y `light-novels`. El estado admite `to-read`, `up-next`, `reading`, `read`, `paused`, `did-not-finish`. Formatos de edición: cadena vacía (desconocido), `paperback`, `hardcover`, `digital`, `audio`, `other`.

Las fechas de biblioteca/consumo son `AAAA-MM-DD` o cadena vacía cuando se desconocen; no tienen hora ni zona horaria. `favorite` y `owned` son booleanos independientes del estado de lectura. `position` es un número no negativo y solo tiene sentido con serie. `identifier` es texto opcional de edición, no se presupone que sea siempre un ISBN.

`identifiers` conserva esquema, valor normalizado, fuente y marca de equivalencia derivada. `catalog` contiene el snapshot bibliográfico: proveedor principal, fecha de consulta RFC 3339, obra y edición remotas, identificadores, descripción, URL de portada, fuentes por campo y advertencias. No sustituye los campos personales. Los bytes de portadas y las credenciales no forman parte del respaldo; las URLs y la metadata sí.

Relaciones: `works.seriesId` → `series.id`; `editions.workId` → `works.id`; `cycles.editionId` → `editions.id`. Una obra y su serie deben pertenecer al mismo medio. Cada obra debe tener edición y registro de lectura. Se conservan todas las ediciones y ciclos, no solo los que seleccione una vista del frontend. Los nombres de serie no pueden colisionar al normalizarlos dentro de un medio, pero pueden repetirse entre medios distintos.

Se validan identidad de formato, versión, duplicados, referencias, fechas, enums y coherencia temporal, sin inventar fechas faltantes. El catálogo y las cuatro colecciones deben estar presentes, aunque sean arrays vacíos: un contenido omitido no se interpreta como una biblioteca vacía. Los campos o medios desconocidos se rechazan, en lugar de descartarlos silenciosamente. La incorporación de nuevos datos debe actualizar explícitamente el formato/compatibilidad y sus pruebas.

Ejemplo **sintético** abreviado de una obra antigua con fechas de consumo desconocidas:

```json
{
  "format": "trove-library",
  "version": 3,
  "generation": "synthetic-generation",
  "createdAt": "2026-01-01T00:00:00Z",
  "catalog": {
    "series": [],
    "works": [{"id":"work-example","medium":"books","title":"Una historia de ejemplo","authors":["Autora ficticia"],"notes":"Nota de ejemplo","addedOn":"","favorite":false}],
    "editions": [{"id":"edition-example","workId":"work-example","language":"en","format":"digital","identifier":"","owned":true}],
    "cycles": [{"id":"cycle-example","editionId":"edition-example","status":"read","startedOn":"","finishedOn":""}]
  }
}
```

Este archivo no contiene SQL ni tipos del driver. Otro motor puede reconstruir sus entidades y relaciones mediante un adaptador nuevo, sin leer una base SQLite primero. No significa conversión automática a cualquier aplicación o motor.

## Compatibilidad con formatos anteriores

Las copias de formato 1 contienen solo `works.medium = books`; sus series no tienen campo `medium` y se interpretan como Libros. El formato 2 incorpora ambos medios, sin snapshots bibliográficos. Se conservan IDs, relaciones, fechas desconocidas y el hash lógico original: validar una copia antigua no la reescribe como formato 3 ni invalida su manifiesto.

Los nuevos respaldos se escriben en formato 3. Las restauraciones antiguas se preparan sobre el esquema actual, sin aplicar fechas de lectura en vivo. Se admiten snapshots nativos identificados con esquemas SQLite 1, 2, 3 o 4 y formato compatible. Una aplicación anterior puede no admitir el esquema 4 o el formato 3.

Restaurar una copia antigua es recuperar ese estado completo, que no contiene novelas ligeras: no conserva por mezcla las entradas añadidas después. La copia previa al reemplazo sí conserva el estado actual completo con ambos medios.

## Respaldo nativo

`trove.sqlite` se genera con un snapshot consistente de SQLite y lleva una tabla técnica `trove_backup` que identifica formato, versión, generación y fecha. No es la base activa ni una copia de sus archivos tomada mientras se escriben. Para leerlo hace falta soporte SQLite.

El JSON se obtiene del snapshot terminado, no de otra consulta posterior a la biblioteca activa. Así ambas representaciones corresponden al mismo estado, aunque aparezcan nuevas escrituras mientras se prepara el respaldo.

## Restauración y límites

Se admiten archivos regulares de hasta 256 MiB por representación. La app prepara una base nueva, verifica que el contenido recuperado coincide y muestra el perfil de destino antes de reemplazar. No mezcla bibliotecas. La selección de archivos usa diálogos nativos.

La configuración de carpeta vive en `backup-settings.json` junto a los datos del perfil y **no** se restaura desde el respaldo. Tampoco se incluyen rutas de máquina, tokens, logs, cachés del WebView o portadas. No se consulta Internet ni se aplican reglas de lectura en vivo al recuperar fechas históricas.

Las copias `before-restore-*.sqlite` anteriores a un reemplazo quedan locales para recuperación; su retirada manual es independiente de la retención automática de generaciones. El usuario debe cerrar otras instancias antes de restaurar. Ante un fallo de reemplazo se intenta reabrir la biblioteca anterior y se informa la ubicación de seguridad si requiere intervención.
