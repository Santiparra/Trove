# Instrucciones para trabajar en Trove

- El proyecto está en planificación. El usuario autorizó crear y mantener documentación; no implementar código, instalar dependencias ni preparar infraestructura sin una nueva instrucción.
- Escribir documentación en español. Trove es el nombre provisional.
- Consultar `docs/decisions.md` antes de asumir decisiones de producto o tecnología.
- Mantener todos los cambios locales. No crear commits, hacer push ni publicar; el usuario gestiona esas acciones. No alterar el staging del usuario.
- Consultar al usuario antes de resolver una decisión importante de producto no cubierta por los acuerdos existentes. Registrar alternativas como propuestas, no como decisiones aprobadas.
- Distinguir requisitos confirmados por el usuario de recomendaciones y asuntos pendientes. Una recomendación no equivale a aprobación.
- Registrar decisiones acordadas y sus consecuencias en los documentos pertinentes; evitar duplicar especificaciones en varios archivos.
- Mantener la arquitectura proporcional a una aplicación personal local para una sola PC.
- Los datos personales deben ser independientes de los proveedores. Importadores y proveedores de metadata son responsabilidades diferentes.
- No integrar IA para recomendaciones: el alcance pedido es exportar contexto para copiar y pegar en herramientas externas.
- Verificar cambios documentales mediante el diff y la coherencia de enlaces y decisiones; no requiere builds ni tests de aplicación.
