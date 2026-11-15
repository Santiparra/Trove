# Roadmap propuesto

No es un compromiso de implementación. El orden puede ajustarse durante la planificación.

| Etapa | Resultado y condición de aceptación propuesta |
| --- | --- |
| 0. Definiciones iniciales | Libros/The StoryGraph y SQLite elegidos; concretar ejemplos reales, unidad de registro, estados, resguardo y stack |
| 1. Biblioteca manual local | Crear, editar y consultar entradas del primer medio, con datos personales persistidos al cerrar y abrir |
| 2. Resguardo y recuperación | Generar automáticamente respaldo ligero y fallback abierto en carpeta configurable; restaurar datos propios sin metadata/portadas y permitir recuperarlas después |
| 3. Primera importación | Rescatar un archivo real con revisión previa y repetición sin duplicados |
| 4. Primer proveedor | Buscar y vincular metadata, conservar snapshot y usar biblioteca sin conexión |
| 5. Actualización | Refrescar metadata sin perder cambios personales ni el último snapshot válido ante fallos |
| 6. Segundo medio | Añadir sus particularidades y comprobar qué conceptos pueden compartirse |
| 7. Exportaciones de uso | Texto de consumo reciente y HTML estático de recomendaciones seleccionadas |
| 8. Estadísticas | Métricas e información por género a partir del historial disponible |

Historial, fechas, ediciones y relaciones deben modelarse cuando el primer recorrido los necesite; no posponer su captura hasta la etapa de gráficos si eso pierde información. Tampoco se implementa por adelantado todo el modelo de todos los medios.

La primera importación será libros desde The StoryGraph (más de 250 ítems según el usuario). Falta examinar una exportación real para definir campos y casos de conflicto.
