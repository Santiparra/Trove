# Formatos de importación y criterios de interpretación

Referencia técnica basada en los formatos examinados, sin cantidades, estadísticas, títulos, identificadores concretos ni historial de una biblioteca personal. Los originales están en `samples/private/`, excluidos de Git. No se han importado datos a una base de Trove ni se han implementado importadores.

Las fuentes de importación no determinan los proveedores de metadata. Este documento describe cómo interpretar campos y qué limitaciones considerar; no es una especificación exhaustiva de cada formato.

## Política de conservación acordada

- Rescatar título, autores, estado, puntaje personal, notas, progreso, datos de consumo y fechas disponibles, además de referencias externas útiles para identificar obras y ediciones.
- Las fechas faltantes son válidas: no inventarlas ni bloquear la importación por su ausencia. Trove permitirá registrarlas y editarlas para el consumo futuro.
- Conservar procedencia e identificadores del origen. No exigir que una obra tenga ISBN o que ya pueda vincularse con un proveedor.
- No incorporar metadata secundaria del archivo como snapshots: títulos alternativos, géneros, ratings globales, sinopsis y demás información descriptiva se obtendrán del proveedor elegido al implementar el enriquecimiento.
- Esta exclusión no afecta al título y autor que identifican el registro propio ni al progreso personal; tampoco cambia por sí sola la política de backup de snapshots que Trove obtenga posteriormente.
- Descartar perfil de cuenta, red social, likes y ajustes propios del servicio.
- La retención la implementará la aplicación en una entrega futura. Analizar archivos no autoriza transformar o persistir ahora una biblioteca real.

## Gestor de importación

Antes de aplicar cambios, el usuario podrá revisar las entradas, filtrarlas, aprobarlas, ignorarlas y asignar su medio de destino individualmente o en bloque. Esto permite separar libros y novelas ligeras de un mismo archivo, o manga, manhwa y novelas ligeras de un XML común.

La clasificación manual no depende de conseguir metadata online. Los IDs externos se mantienen aunque se cambie el medio de destino. Los detalles visuales y reglas de aplicación se definirán al implementar el gestor; las políticas de duplicados están en [integraciones](integrations.md).

## The StoryGraph

CSV con campos bibliográficos, personales y de valoración. Usar un lector CSV que respete comillas, separadores y texto; no dividir líneas o autores ingenuamente por comas.

| Campos | Uso previsto |
| --- | --- |
| `Title`, `Authors` | Identificación local editable |
| `ISBN/UID` | Referencia externa opcional, almacenada como texto; validar su esquema antes de tratarla como ISBN |
| `Contributors`, `Format` | Posibles datos propios de edición; mapeo preciso pendiente, sin convertirlos en requisito |
| `Read Status` | Traducir estados como `read`, `paused`, `to-read`, `did-not-finish`, `currently-reading` al modelo local |
| `Date Added` | Fecha de agregado al origen, no de importación ni de lectura |
| `Last Date Read`, `Dates Read`, `Read Count` | Fechas y repeticiones disponibles, sin inventar ciclos detallados a partir de un contador |
| `Star Rating`, `Review` | Puntaje y reseña personales |
| `Tags`, `Owned?` | Datos personales opcionales; alcance preciso pendiente |

El campo `ISBN/UID` puede estar vacío o contener esquemas distintos. Una edición agregada por usuarios puede estar incompleta; el título y los autores siguen siendo útiles sin completar toda su ficha.

Las fechas del formato pueden usar `AAAA/MM/DD` e intervalos `AAAA/MM/DD-AAAA/MM/DD`. Una cuenta creada recientemente puede incorporar consumo antiguo sin fechas; la fecha de agregado no permite reconstruir ese consumo.

El archivo no tiene campos dedicados de serie, posición en serie o relaciones entre ediciones. Detectar un volumen por texto del título es una sugerencia revisable, no una vinculación fiable.

`Star Rating` admite decimales. La escala y precisión implementadas se definirán por medio teniendo en cuenta el proveedor elegido y el mapeo desde cada origen; no fijar una escala universal ni redondear silenciosamente durante la importación.

Campos como `Moods`, `Pace`, preguntas sobre personajes y advertencias de contenido no se incorporan automáticamente como metadata ni como funcionalidades de Trove. Si posteriormente se quieren rescatar como valoraciones personales adicionales, se definirá expresamente su mapeo.

## MyAnimeList: anime

XML con raíz `myanimelist`, bloque `myinfo` y elementos `anime`; puede recibirse como XML o comprimido gzip. Los datos de cuenta de `myinfo` no forman parte de la biblioteca a importar.

| Campos | Uso previsto |
| --- | --- |
| `series_animedb_id`, `series_title` | Referencia MAL anime y título local |
| `my_status`, `my_watched_episodes` | Estado y progreso personal |
| `my_score`, `my_comments`, `my_tags` | Puntaje, comentarios y etiquetas personales según alcance |
| `my_start_date`, `my_finish_date` | Fechas conocidas; `0000-00-00` significa desconocido |
| `my_times_watched`, `my_rewatching`, `my_rewatching_ep` | Información de repetición que requiere interpretar las convenciones del origen |

`series_type` y `series_episodes` describen formato y total de episodios; pueden ayudar a presentar la revisión, pero no se persisten como snapshots importados. El progreso personal no debe descartarse por no conservar el total descriptivo.

Un total cero no debe asumirse como total real; un `my_score` cero debe interpretarse según el contrato del origen, diferenciándolo de una puntuación personal efectiva. El campo `my_id` no sustituye a `series_animedb_id` como referencia del catálogo.

`update_on_import` pertenece al formato MAL y no decide la política de Trove. Los formatos `TV`, `Movie`, `OVA`, `ONA` o especiales no son por sí solos secciones de destino: una película de anime puede pertenecer al medio anime.

No deducir secuelas, precuelas ni ciclos completos a partir de títulos o contadores sin información suficiente.

## MyAnimeList: manga, manhwa y novelas ligeras

El XML contiene elementos `manga`, con referencia externa `manga_mangadb_id` y título `manga_title`. **El ID existe**, pero el archivo no aporta un campo explícito fiable para distinguir los medios. Conocer un ID de catálogo no equivale a conocer su subtipo sin consultar información adicional.

El gestor permitirá filtrar, seleccionar en bloque, asignar el medio de destino e ignorar entradas. No es necesario consultar una API para permitir esa selección manual.

`my_read_chapters` y `my_read_volumes` son progreso personal y deben mantenerse diferenciados. `manga_chapters` y `manga_volumes` son totales descriptivos externos, no progreso. Estados, puntajes, comentarios, tags y fechas siguen criterios equivalentes al importador anime, con sus campos propios.

Una entrada agregada de serie no equivale a sus volúmenes individuales. No fabricar ediciones o ciclos por volumen a partir de un contador; definir el mapeo al incorporar este importador.

## Trakt

ZIP con varios JSON y posibles fragmentos numerados. Procesar cada conjunto por su estructura y reunir sus fragmentos sin depender del orden lexicográfico de archivos.

| Conjunto | Uso previsto |
| --- | --- |
| `watched-shows-*.json`, `watched-movies.json` | Resúmenes personales de reproducciones y última visualización |
| `watched-history-*.json` | Eventos de consumo con identificador y fecha/hora |
| `ratings-shows.json`, `ratings-movies.json` | Puntajes personales de serie/película y fecha de calificación |
| `ratings-seasons.json`, `ratings-episodes.json` | Puntajes personales con relación a su temporada/episodio y serie |
| `lists-watchlist.json`, archivos de listas | Pertenencia, orden y fechas, según alcance de la feature |

Los mapas de IDs pueden incluir `trakt`, `imdb`, `tmdb` y `tvdb`, según entidad. Guardar namespace y tipo junto al identificador. La posición de un episodio y su serie padre son referencias estructurales necesarias para asociar su progreso, no un snapshot completo de catálogo.

El historial distingue eventos de obras: repetir un episodio no duplica la obra. No sumar resúmenes e historial como reproducciones independientes. Los resúmenes pueden no incluir el detalle de episodios; el historial permite recuperarlo cuando está disponible.

Los timestamps pueden estar en UTC con sufijo `Z`; no confundirlos con fechas sin hora. Ver un episodio no demuestra completar una serie y los eventos no definen automáticamente ciclos completos de revisualización.

Ocultar una entrada de recomendaciones/progreso no equivale a abandonarla. Perfil, red social, likes y configuración se descartan. Reglas finas de importación se resolverán al implementar Trakt.

## IMDb: fuente posterior de puntajes

CSV de calificaciones. Se incorporará después de Trakt, para usar sus puntuaciones personales como fuente de verdad cuando el usuario seleccione sobrescribir los valores existentes; no es una importación prioritaria ni un proveedor de metadata aprobado.

- `Const`: referencia IMDb para emparejar el registro.
- `Your Rating`: puntaje personal; no confundir con `IMDb Rating`.
- `Date Rated`: fecha de puntuación, no de consumo.
- Título, tipo y otros campos pueden ayudar temporalmente a revisar coincidencias, sin guardar metadata secundaria del archivo.
- `IMDb Rating`, `Num Votes`, géneros y demás datos descriptivos no se incorporan como snapshots.

Los encabezados y los valores de categorías pueden tener idiomas distintos. Una categoría general de serie no distingue necesariamente anime de otros contenidos.

La finalidad es emparejar y actualizar calificaciones, no reconstruir estados ni generar una biblioteca nueva a partir de ratings. El tratamiento de calificaciones sin coincidencia se definirá en esa feature; no asumir creación automática de entradas.

## Consecuencias para el diseño

1. Cada origen tiene su adaptador; normalización, revisión manual, coincidencias y aplicación de cambios son responsabilidades distintas.
2. El usuario decide medio de destino y registros que acepta; la metadata externa se vincula en un recorrido separado.
3. Un mismo identificador externo puede ayudar a cruzar fuentes, pero no justifica fusionar niveles distintos (obra, edición, serie, temporada o episodio).
4. Las fechas desconocidas se mantienen vacías. Trove sí debe permitir registrar fechas fiables para consumo futuro.
5. Mantener referencias externas e información personal sin copiar todo el payload del origen.
6. Los datos reales y sus estadísticas no son documentación pública. Los ejemplos que se agreguen al repo deben ser sintéticos y no derivados de perfiles personales identificables.
