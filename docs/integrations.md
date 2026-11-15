# Importaciones, metadata y exportaciones

## Distinción confirmada

Un servicio del que se importa información personal no tiene por qué ser proveedor de metadata de Trove. Por ejemplo, rescatar una lista de MyAnimeList no obliga a usarlo como motor de búsqueda o actualización.

## Orígenes previstos

MyAnimeList, The StoryGraph y Trakt para listados; IMDb para puntajes. Formatos exportables, APIs, autenticación y campos reales pendientes de verificar con archivos o documentación. No se presupone un método de extracción disponible para cada servicio.

## Flujo recomendado de importación

Leer archivo → normalizar sin perder procedencia → previsualizar altas/cambios/conflictos → confirmar → persistir datos propios → enriquecer metadata posteriormente si se desea.

Objetivos propuestos: repetir una importación sin duplicar; conservar IDs originales; no sobrescribir notas o cambios posteriores silenciosamente; preservar fechas desconocidas como desconocidas; conservar escala original de puntajes cuando sea necesario.

## Proveedores de metadata

Todos por definir. Evaluar cobertura por medio, granularidad de entidades, búsqueda, detalles, imágenes, disponibilidad de IDs, límites y necesidad de cuenta. Una API puede no cubrir todas las capacidades.

La vinculación debe permitir revisión manual y corrección. Los snapshots deben registrar proveedor, ID y fecha, y conservarse aunque falle el origen.

## Salidas solicitadas

- **Contexto para IA externa:** texto copiable con las últimas N obras consumidas y campos a elegir. No hay integración de IA.
- **Recomendaciones para amigos:** selección manual de obras → HTML estático con información de cada una. Un HTML no implica publicación ni hosting automático.

Pendiente para HTML: archivo único o paquete con imágenes; uso offline; selección de notas y puntajes personales; orden y presentación; metadatos incluidos. No exponer notas privadas por defecto sin acordar esa regla con el usuario.
