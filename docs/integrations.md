# Importaciones, metadata y exportaciones

## Distinción confirmada

Un servicio del que se importa información personal no tiene por qué ser proveedor de metadata de Trove. Por ejemplo, rescatar una lista de MyAnimeList no obliga a usarlo como motor de búsqueda o actualización.

## Orígenes previstos

Se examinaron los formatos aportados por el usuario. Ver [formatos y criterios de interpretación](import-analysis.md) para campos y límites de información, sin estadísticas personales. Son fuentes de importación, no proveedores de metadata elegidos.

MyAnimeList, The StoryGraph y Trakt para listados; IMDb para puntajes. Formatos exportables, APIs, autenticación y campos reales pendientes de verificar con archivos o documentación. No se presupone un método de extracción disponible para cada servicio.

IMDb se incorporará después de Trakt y solo para emparejar/actualizar calificaciones personales, con prioridad del origen cuando el usuario la seleccione. No reconstruye historial de consumo.

Alcance confirmado: el usuario obtiene las exportaciones fuera de Trove; la app solo importa los ficheros ya generados. Los ejemplos privados para análisis se pueden dejar en `samples/private/`, excluidos de Git salvo su README. No convertirlos en fixtures públicos sin preparar una versión anonimizada acordada.

## Contexto de ejecución confirmado

Cada pestaña representa un medio con tipo interno propio y un proveedor de metadata vinculado, todavía por elegir. El listado y el medio determinan el alcance y el adaptador utilizado. La cobertura adicional del proveedor no amplía la operación: refrescar anime no consulta ni modifica libros o películas.

Importadores de origen y adaptadores de metadata son responsabilidades distintas, aunque un servicio pueda participar en ambos roles.

## 1. Buscar y consultar información

La búsqueda externa permite ver candidatos y detalles sin incorporarlos a la biblioteca. Solo la acción explícita de agregar a un listado crea el registro de biblioteca correspondiente. La búsqueda por sí sola no genera contenido para backups.

Puede existir caché temporal, excluida del respaldo. Duración y limpieza pendientes; no se exige persistir resultados descartados. Consultar un resultado tampoco debe crear automáticamente una vinculación con una obra local existente.

## 2. Importar datos personales

Operación sensible: cada origen tiene formato, estructura y campos propios, y necesita un adaptador de importación específico. Se rescatan datos propios (títulos, autores, estado, puntaje, notas, progreso, fechas disponibles) y referencias externas. No se guardan como snapshots la metadata secundaria del archivo —títulos alternativos, géneros, ratings globales y demás información descriptiva—; ese enriquecimiento corresponde al proveedor elegido posteriormente. Perfil, red social, likes y ajustes del servicio se descartan.

Leer archivo → normalizar sin perder procedencia → previsualizar altas/cambios/conflictos → confirmar → persistir datos propios → enriquecer metadata posteriormente si se desea.

Gestor confirmado: revisar, filtrar, aprobar, ignorar y asignar medio de destino por entrada o en bloque antes de aplicar cambios. Permite separar libros y novelas ligeras de StoryGraph, o clasificar manga/manhwa/novelas ligeras en MAL, conservando sus IDs externos. El ID de MAL identifica la entrada, pero no aporta por sí solo su clasificación en el XML.

Fechas faltantes no bloquean el rescate: no se inventan. Trove ofrecerá registro y edición de fechas para consumo futuro. Escalas y decimales se concretan por medio considerando el proveedor y el origen; no se redondean silenciosamente.

Objetivos propuestos: repetir una importación sin duplicar; conservar IDs originales; no sobrescribir notas o cambios posteriores silenciosamente; preservar fechas desconocidas como desconocidas; conservar escala original de puntajes cuando sea necesario.

Requisito confirmado: emparejar entradas entrantes con registros locales y ofrecer opciones de actualización de duplicados. Este emparejamiento evita duplicados locales y es distinto de vincular la obra con un proveedor de metadata. La importación debe poder generar el listado local antes de enriquecerlo desde proveedores.

Política confirmada, configurable al importar:

- **Conservar existentes (prioridad local):** si la entrada ya existe en el listado local, no modificarla. No completar ni actualizar campos de ese registro mediante esta importación.
- **Sobrescribir existentes (prioridad del origen):** si la entrada ya existe, aplicar los datos entrantes sobre los campos locales correspondientes según el mapeo del importador, aunque hayan sido editados localmente.

Esta selección determina la fuente de verdad para las coincidencias de esa importación; no requiere comparar con una importación anterior. No implica reemplazar la identidad interna ni convertir la importación en un refresco de metadata.

Ejemplo: puntaje local 5 y puntaje entrante 4 → conservar mantiene 5; sobrescribir lo cambia a 4. La regla de identificación de coincidencias se define aparte: «ya existe» no equivale a aprobar coincidencia solo por texto del título, especialmente entre ediciones.

Pendiente: campos incluidos en cada mapeo, distinción entre campo ausente y valor vacío explícito, coincidencias ambiguas y tratamiento de entradas ausentes del archivo. No asumir borrado por ausencia ni limpieza de campos no proporcionados.

## 3. Refrescar o reconstruir metadata

Operación explícita y repetible sobre entradas ya existentes del listado interno. Utiliza el tipo de medio, las vinculaciones disponibles y su adaptador para obtener metadata externa y actualizar snapshots o recuperar recursos ausentes.

Sirve tanto para actualizaciones ocasionales como para reconstrucción después de restaurar un respaldo. No vuelve a importar el historial personal ni debe crear duplicados o sobrescribir estados, notas, fechas y progreso. No toca otros medios.

Si falta una vinculación válida, se necesita buscar candidatos y emparejar según las reglas de revisión acordadas; no asumir que el identificador del archivo importado sirve directamente en el proveedor de metadata elegido. Repetir un refresco mantiene la identidad local, aunque la información externa pueda cambiar entre consultas.

## Proveedores de metadata

Todos por definir. Evaluar cobertura por medio, granularidad de entidades, búsqueda, detalles, imágenes, disponibilidad de IDs, límites y necesidad de cuenta. Una API puede no cubrir todas las capacidades.

La vinculación debe permitir revisión manual y corrección. Los snapshots deben registrar proveedor, ID y fecha, y conservarse aunque falle el origen.

Obras y ediciones manuales pueden quedar sin vincular indefinidamente. Nombre propuesto para su enriquecimiento posterior: **vinculación de metadata**; busca candidatos y asocia la entrada externa al registro local existente, sin exigir eliminar la edición manual ni sustituir su identidad o sus datos personales. Nombre definitivo de interfaz pendiente.

## Salidas solicitadas

- **Contexto para IA externa:** texto copiable con las últimas N obras consumidas y campos a elegir. No hay integración de IA.
- **Recomendaciones para amigos:** selección manual de obras → HTML estático con información de cada una. Un HTML no implica publicación ni hosting automático.

Pendiente para HTML: archivo único o paquete con imágenes; uso offline; selección de notas y puntajes personales; orden y presentación; metadatos incluidos. No exponer notas privadas por defecto sin acordar esa regla con el usuario.
