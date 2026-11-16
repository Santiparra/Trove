# Roadmap propuesto

La primera entrega está acordada. Las etapas posteriores siguen siendo orientativas y no autorizan implementación automática.

## Primera entrega acordada: navegación y libros locales

- Ventana principal con secciones Libros, Novelas ligeras, Anime, Manga, Manhwa, Series de TV y Películas. Solo Libros funcional; el resto muestra «En construcción».
- Biblioteca manual de libros: listado, creación, detalle, edición y búsqueda local por título.
- Serie opcional compartida y número de orden dentro de ella para vincular libros serializados.
- Campos iniciales propuestos y aceptados: título, autor/autores, edición mínima con idioma/formato/ISBN opcionales, estado, notas y fechas de agregado/inicio/finalización cuando se conozcan. Puntaje condicionado a acordar su escala antes de implementarlo; no elegirla silenciosamente.
- Conexión real React → puente Wails → casos de uso Go → SQLite, con inicialización y esquema versionado. Guardado y errores visibles, sin fingir éxito ante fallos.
- Diferenciar obra, edición y ciclo de consumo desde el modelo. La interfaz inicial trabaja con una edición por libro; no incluye gestión avanzada de ediciones o relecturas.
- Directorios de datos de desarrollo/prueba separados del uso habitual. Datos reemplazables para esta entrega.
- Comprobaciones focalizadas de reglas y persistencia, más prueba del recorrido en paquete portable local. Esto no asigna una versión ni dispara una release.

Fuera de esta entrega: importaciones, proveedores y búsqueda externa, vinculación, portadas/snapshots, gestión avanzada de series/ediciones, estadísticas, recomendaciones y sistema automático de backup/restauración. Este último se abordará antes de depender de Trove como único registro de información importante.

Aceptación: navegar por todas las secciones; crear un libro sin ISBN ni proveedor; encontrarlo por título; editarlo; cerrar completamente la app y recuperar los cambios al reabrir. Comprobar también desde una build portable de prueba sin servidor de desarrollo.

## Etapas posteriores y contexto

| Etapa | Resultado y condición de aceptación propuesta |
| --- | --- |
| 0. Definiciones iniciales | Libros/The StoryGraph y SQLite elegidos; concretar ejemplos reales, unidad de registro, estados, resguardo y stack |
| 1. Biblioteca manual local | Crear, editar y consultar entradas del primer medio, con datos personales persistidos al cerrar y abrir |
| 2. Resguardo y recuperación | Generar automáticamente respaldo ligero y fallback abierto en carpeta configurable; restaurar datos propios y contenido textual incluido, sin portadas/recursos pesados, y permitir recuperarlos después |
| 3. Primera importación | Rescatar un archivo real con revisión previa y repetición sin duplicados |
| 4. Primer proveedor | Buscar y vincular metadata, conservar snapshot y usar biblioteca sin conexión |
| 5. Actualización | Refrescar o reconstruir metadata por medio/proveedor sin perder cambios personales, sin modificar otros medios y conservando el último snapshot válido ante fallos |
| 6. Segundo medio | Añadir sus particularidades y comprobar qué conceptos pueden compartirse |
| 7. Exportaciones de uso | Texto de consumo reciente y HTML estático de recomendaciones seleccionadas |
| 8. Estadísticas | Métricas e información por género a partir del historial disponible |

Historial, fechas, ediciones y relaciones deben modelarse cuando el primer recorrido los necesite; no posponer su captura hasta la etapa de gráficos si eso pierde información. Tampoco se implementa por adelantado todo el modelo de todos los medios.

La primera importación será libros desde The StoryGraph; ver [formatos y criterios de interpretación](import-analysis.md). El gestor incluirá revisión y clasificación manual en bloque. Las fechas faltantes se aceptan y el mapeo exacto se concreta al preparar la feature. IMDb queda para una etapa posterior a Trakt y solo como fuente de calificaciones personales.
