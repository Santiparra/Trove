# Arquitectura propuesta

Estado: propuesta pendiente. No define todavía frameworks, estructura de carpetas ni contratos de código.

## Restricciones confirmadas

Aplicación local para una sola PC. El usuario abre la app y no administra servicios. La gestión personal funciona sin proveedores; las integraciones enriquecen y rescatan datos.

## Organización recomendada

1. **Interfaz:** biblioteca, edición, progreso, importaciones y vinculación.
2. **Casos de uso:** acciones completas como agregar obra, terminar lectura, importar archivo o actualizar metadata.
3. **Dominio:** reglas e identidad propias, sin dependencia de APIs externas.
4. **Adaptadores:** persistencia, archivos, importadores y proveedores de metadata.

Un único producto modular es suficiente. Separar frontend y backend conceptualmente no exige dos aplicaciones, un servidor web ni microservicios.

## Límites importantes

- IDs de proveedores no sustituyen la identidad interna propuesta.
- Importar un historial no exige encontrar una obra en un proveedor de metadata.
- Un fallo de actualización no debe eliminar el último snapshot válido.
- Los datos personales y sus modificaciones deben tener reglas explícitas de precedencia.
- Las capacidades de proveedores pueden diferir; no asumir intercambiabilidad perfecta.
- Las credenciales deben permanecer fuera de exportaciones compartidas; su almacenamiento depende del stack elegido.

## Decisiones pendientes

Contenedor de escritorio, lenguaje del núcleo, persistencia, canal de comunicación UI/núcleo, distribución y ubicación de datos. Ver [stack](tech-stack.md) y [backend](backend.md).
