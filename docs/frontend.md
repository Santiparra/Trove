# Frontend

Estado inicial implementado: navegación por medios, biblioteca local de libros, búsqueda por título, detalle y formulario de edición con Serie y número. React se comunica con los casos de uso Go mediante bindings generados de Wails. Las restantes secciones muestran «En construcción».

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

Campos mínimos, estados y escala de puntaje de libros (primer medio elegido); detalles de navegación; densidad visual; selección de edición; presentación de series; contenido predeterminado de exportaciones. No se generan mockups ni código en esta etapa.
