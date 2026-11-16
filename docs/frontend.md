# Frontend

Estado implementado: navegación por medios con iconos SVG, portal de libros, listados por estado, búsqueda por título, ficha y formulario a página completa. React se comunica con los casos de uso Go mediante bindings generados de Wails. Las restantes secciones muestran «En construcción».

## Portal y navegación de Libros

- Libros abre el portal: estados con cantidades, lecturas actuales, últimos libros terminados con fecha y barra proporcional por estado.
- Elegir un estado abre su listado; hay acceso a todos y la búsqueda por título también abre el listado. Se conserva el filtro al volver desde una ficha.
- Ficha y formulario ocupan el área principal completa, sin tapar la barra lateral. El formulario distribuye identificación, edición, lectura y notas en secciones.
- El listado y las lecturas actuales permiten cambiar estado sin editar toda la ficha. La acción actualiza únicamente el estado, conserva las fechas existentes y no inventa una finalización.
- Los últimos leídos se ordenan por fecha de finalización conocida y se limitan a cinco. Los leídos sin fecha siguen en sus listados y se informa su existencia, sin atribuirles un orden cronológico ficticio.
- Hasta integrar portadas, se muestran marcadores locales claramente identificados como «Sin portada», sin tráfico externo ni imágenes inventadas de obras.
- Eliminar está disponible en listado y ficha, con confirmación explícita. Se eliminan libro, ediciones y ciclos asociados de esta biblioteca, no otros libros ni su serie compartida. Si falla, se conserva el registro y se muestra el error.
- Datos, filtros, recuentos y operaciones autoritativas siguen en Go; React representa distribución y navegación. No requiere migración de esquema respecto de la primera entrega.

## Dirección confirmada

Experiencia de escritorio para una PC. Secciones especializadas por medio con listados personales, búsqueda, filtros, notas, puntajes, estados y progreso. No hay diseño visual aprobado.

Cada medio dispone de una pestaña separada y de su proveedor vinculado. Tres operaciones diferenciadas: buscar/consultar sin guardar en biblioteca; importar datos personales con emparejamiento y políticas de duplicados; refrescar/reconstruir metadata del listado existente. Sus reglas están en [integraciones](integrations.md).

## Recorridos a definir

- Crear y editar una entrada manualmente.
- Consultar biblioteca y actualizar consumo individual y progreso de serie.
- Ver obra, edición, relaciones y metadata, diferenciando información personal de externa.
- Previsualizar una importación y resolver conflictos.
- Gestor de importación con filtros, aprobación, ignorar y asignación de medio de destino individual o en bloque antes de guardar; permite separar contenido mezclado sin consulta online obligatoria.
- Revisar candidatos de metadata y vincular, rechazar o corregir coincidencias.
- Actualizar snapshots y mostrar última actualización o fallo sin perder datos.
- Exportar/restaurar datos y mostrar dónde están guardados.
- Copiar un listado reciente como texto.
- Seleccionar obras y generar un HTML estático para compartir.

## Propuestas de diseño

Compartir componentes comunes y especializar lo necesario por medio. Distinguir búsqueda local de búsqueda externa. Mostrar las relaciones sin fusionar obras ni duplicar consumo.

## Pendiente

Escala de puntaje, portadas y metadata, gestor avanzado de ediciones/series, ajustes visuales según uso y contenido predeterminado de exportaciones. Los estados y campos iniciales ya están implementados; no se anticipan las features posteriores.
