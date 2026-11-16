# Instrucciones para trabajar en Trove

- Primera entrega autorizada en Principal, en este chat: navegación por medios y biblioteca manual de libros con SQLite, búsqueda local, Serie y número de orden. Pulido solicitado: portal compacto, fichas/formularios completos, cambio rápido, eliminación confirmada, iconos, navegación de serie, búsqueda desplegable, notas en ficha, favorito/propiedad, A continuación y fechas locales automáticas. Ver `docs/roadmap.md`. No ampliar el alcance sin acuerdo.
- Escribir documentación en español. Trove es el nombre provisional.
- Consultar `docs/decisions.md` antes de asumir decisiones de producto o tecnología.
- Mantener todos los cambios locales. No crear commits, hacer push ni publicar; el usuario gestiona esas acciones. No alterar el staging del usuario.
- Consultar al usuario antes de resolver una decisión importante de producto no cubierta por los acuerdos existentes. Registrar alternativas como propuestas, no como decisiones aprobadas.
- Distinguir requisitos confirmados por el usuario de recomendaciones y asuntos pendientes. Una recomendación no equivale a aprobación.
- Registrar decisiones acordadas y sus consecuencias en los documentos pertinentes; evitar duplicar especificaciones en varios archivos.
- Mantener la arquitectura proporcional a una aplicación personal local para una sola PC.
- Mantener frontend centrado en presentación e interacción y lógica de negocio autoritativa en backend. Desacoplar núcleo de Wails, proveedores y SQLite mediante límites concretos; evitar abstracciones genéricas innecesarias y duplicación de reglas. Ver `docs/architecture.md`.
- Los datos personales deben ser independientes de los proveedores. Importadores y proveedores de metadata son responsabilidades diferentes.
- No incluir cantidades, estadísticas, títulos ni identificadores de bibliotecas personales en documentación pública. Describir formatos con ejemplos sintéticos; mantener exportaciones originales excluidas de Git.
- No integrar IA para recomendaciones: el alcance pedido es exportar contexto para copiar y pegar en herramientas externas.
- El usuario decide cuándo preparar una release. No incrementar versiones, crear notas versionadas, tags ni builds de release automáticamente al terminar una feature. Ver `docs/releases/README.md`.
- Resguardo y restauración autorizados: automático SQLite + JSON lógico, dos generaciones válidas, carpeta configurable por perfil y restauración completa validada con copia previa y recuperación ante fallos. No mezclar con importaciones externas. Ver `docs/backend.md`.
- Verificar cambios documentales mediante el diff y la coherencia de enlaces y decisiones; no requiere builds ni tests de aplicación.
