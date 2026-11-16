# Frontend

Estado implementado: navegación por medios con iconos SVG, portal de libros, listados por estado, búsqueda por título, ficha y formulario a página completa. React se comunica con los casos de uso Go mediante bindings generados de Wails. Las restantes secciones muestran «En construcción».

## Portal y navegación de Libros

- Libros abre el portal: estados con cantidades, lecturas actuales, últimos libros terminados con fecha y barra proporcional por estado.
- Elegir un estado abre su listado; los estados también se recorren mediante botones. La búsqueda muestra un desplegable sin abandonar la vista, con opción de abrir todas las coincidencias. Se conserva el filtro al volver desde una ficha.
- Ficha y formulario ocupan el área principal completa, sin tapar la barra lateral. El formulario distribuye identificación, edición, lectura y notas en secciones.
- El listado y las lecturas actuales permiten cambiar estado sin editar toda la ficha. Se aplican las [reglas de fechas locales](domain.md#registro-personal-y-fechas), conservando los otros datos personales.
- Los últimos leídos se ordenan por fecha de finalización conocida y se limitan a cinco. Los leídos sin fecha siguen en sus listados y se informa su existencia, sin atribuirles un orden cronológico ficticio.
- Hasta integrar portadas, se muestran marcadores locales claramente identificados como «Sin portada», sin tráfico externo ni imágenes inventadas de obras.
- Eliminar está disponible en listado y ficha, con confirmación explícita. Se eliminan libro, ediciones y ciclos asociados de esta biblioteca, no otros libros ni su serie compartida. Si falla, se conserva el registro y se muestra el error.
- Datos, filtros, recuentos y operaciones autoritativas siguen en Go; React representa distribución y navegación. Favorito y propiedad requieren el esquema 2, con copia SQLite consistente previa a migrar una base existente.

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

## Pulido de navegación y registro personal

El bloque de navegación y registro personal está autorizado e implementado. La personalización avanzada queda para una etapa posterior.

### Navegación y uso cotidiano

- Cabecera y controles del portal compactos, con 1920×1080 maximizado como referencia. Layout adaptable mediante altura/anchura disponible y vistas previas acotadas con acceso al listado completo. En ventanas muy bajas se permite reflujo y scroll antes que recortar contenido. La comprobación visual depende del viewport/escalado reales.
- En listados, sustituir el selector de estado por botones visibles centrados arriba, junto a «Volver al portal», para cambiar con un clic.
- Mostrar la serie encima del título, con tamaño cercano al de este pero jerarquía/color propios. Su nombre será clickeable y abrirá una página con los libros locales vinculados, ordenados por su posición en la serie; los ítems sin posición necesitan un orden secundario estable.
- La búsqueda muestra hasta seis coincidencias locales en un desplegable, sin cambiar la página de fondo al escribir. Una acción al pie o Enter abre el listado completo; Escape cierra las sugerencias. No es búsqueda externa ni persiste búsquedas como biblioteca.
- Permitir editar notas desde la ficha, con un botón pequeño «Guardar», sin abrir el formulario completo. El guardado debe actualizar únicamente las notas y conservar el resto de datos.
- Controles de favorito y propiedad en ficha y formulario, y estado manual «A continuación»; reglas en [dominio](domain.md).

### Personalización prevista para después

Reservar conceptualmente una columna a la derecha de lectura actual/últimos leídos para tarjetas más pequeñas: gráfico mensual del año con libros y páginas leídos; portadas de los tres mejor puntuados del año; portadas de favoritos personales de todos los tiempos (cantidad visual por concretar).

No mostrar métricas ficticias ni añadir tarjetas vacías permanentes antes de disponer de sus datos. El ranking anual requiere escala de puntaje, fechas y criterio de desempate; las páginas por mes necesitan definir qué se cuenta y cómo se registra. Favoritos personales no son lo mismo que los mejores puntajes.

Autores clickeables con página propia para descubrir sus obras mediante el proveedor: previsto para la etapa de metadata. No sustituirlo silenciosamente por una búsqueda solo local; se puede añadir navegación local si se acuerda por separado.
