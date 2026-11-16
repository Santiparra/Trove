# Arquitectura propuesta

Estado: arquitectura planteada aceptada por el usuario, con stack Wails v3 + Go + React/TypeScript + Vite + SQLite. Los contratos y la estructura concreta de paquetes se definirán al implementar.

## Restricciones confirmadas

Aplicación local para una sola PC. El usuario abre la app y no administra servicios. La gestión personal funciona sin proveedores; las integraciones enriquecen y rescatan datos.

Dos capas de datos confirmadas: registro local independiente (título, referencias, estado, notas, fechas, progreso) y metadata vinculada a proveedores. Son responsabilidades distintas, no necesariamente dos archivos ni dos bases. El tamaño y el formato, no solo la procedencia, determinan qué entra en el respaldo; ver [persistencia](backend.md).

Los medios deben ser moderadamente independientes y admitir estructuras especializadas. Reconstruir metadata de libros no debe alterar series de TV, incluso si comparten proveedor. El alcance de cada operación debe identificar medio, proveedor y registros objetivo.

## Organización recomendada

1. **Interfaz:** biblioteca, edición, progreso, importaciones y vinculación.
2. **Casos de uso:** acciones completas como agregar obra, terminar lectura, importar archivo o actualizar metadata.
3. **Dominio:** reglas e identidad propias, sin dependencia de APIs externas.
4. **Adaptadores:** persistencia, archivos, importadores y proveedores de metadata.

Un único producto modular es suficiente. Separar frontend y backend conceptualmente no exige dos aplicaciones, un servidor web ni microservicios.

## Principios confirmados y aplicación propuesta

El usuario solicita convenciones consistentes, mínimo código duplicado, frontend ligero, lógica de negocio en backend y persistencia desacoplada. No exige DDD ni una metodología específica.

- **Frontend React:** presentación, navegación, formularios, estado de interacción y validaciones para experiencia de uso. No contiene reglas autoritativas de consumo, importación, emparejamiento, resguardo o SQL. Tener frontend ligero no significa trasladar al backend detalles visuales.
- **Backend Go:** casos de uso y validación autoritativa, reglas de consumo, orquestación de importaciones, metadata y resguardos. Recibir datos desde la UI no evita validarlos nuevamente en esta frontera.
- **Puente Wails:** adaptador fino que transforma entradas/salidas y delega en casos de uso. El dominio no importa Wails.
- **Persistencia:** contratos pequeños según necesidades de los casos de uso; implementación SQLite encapsula SQL, driver y detalles de esquema. El dominio no depende de `sql.DB`, filas SQL ni errores del driver. Los límites transaccionales se definen por operación de negocio, sin filtrar conexiones al núcleo.
- **Integraciones:** tipos de respuesta de APIs y formatos de importación se traducen en los adaptadores; no se convierten en el modelo interno. Cada medio conserva sus reglas y estructura propias.
- **Composición:** conectar implementaciones concretas en el arranque, no crearlas dentro de las reglas de negocio. Preferir composición explícita e idiomática de Go a un framework de inyección.
- **Reutilización:** extraer lo que realmente es común; no forzar repositorios genéricos ni una superentidad multimedia para eliminar similitudes superficiales. Compartir mecanismos, especializar reglas.
- **Convenciones propuestas:** nombres de código en inglés, documentación en español, formato estándar de Go, TypeScript estricto y límites de módulos claros. Aprobar convenciones específicas al iniciar el código.

Dirección propuesta de dependencias: interfaz → adaptador Wails → casos de uso → dominio/contratos; SQLite y proveedores implementan contratos del núcleo. Las implementaciones externas dependen de esos contratos, no al revés.

Migrar de SQLite seguiría requiriendo un nuevo adaptador, migración de datos y comprobaciones; el objetivo es evitar reescribir interfaz y reglas de negocio, no prometer sustitución gratuita de cualquier motor.

## Comprobación arquitectónica al implementar

Propuesta: probar reglas y casos de uso sin arrancar Wails ni consultar proveedores; comprobar persistencia y transacciones contra SQLite real en pruebas focalizadas; probar el mapeo de importaciones con archivos representativos. Mantener pruebas de round-trip de exportación/restauración cuando exista ese recorrido. No construir adaptadores alternativos ficticios ni un framework general para demostrar desacoplamiento.

## Límites importantes

- IDs de proveedores no sustituyen la identidad interna propuesta.
- Importar un historial no exige encontrar una obra en un proveedor de metadata.
- Un fallo de actualización no debe eliminar el último snapshot válido.
- Los datos personales y sus modificaciones deben tener reglas explícitas de precedencia.
- Las capacidades de proveedores pueden diferir; no asumir intercambiabilidad perfecta.
- Las credenciales deben permanecer fuera de exportaciones compartidas; su almacenamiento depende del stack elegido.

## Decisiones pendientes

Implementación inicial: `internal/books` contiene modelo y casos de uso; `internal/storage/sqlite` implementa el contrato de persistencia y migraciones; `internal/desktop` adapta Wails; `internal/config` separa perfiles; `main.go` compone dependencias. React y sus tests están en `frontend/src`, con bindings generados en `frontend/bindings`. El frontend no accede a SQLite ni a proveedores.

`internal/backup` define formato lógico, validación y ciclo de generaciones/restauración mediante un contrato de backend sin importar SQLite. El adaptador de almacenamiento captura/prepara/instala las bases; `internal/desktop/Backups` adapta los diálogos y serializa el reemplazo frente a operaciones de biblioteca. Los guardados confirmados notifican cambios sin depender de que la salida de respaldo esté disponible. Configuración de rutas y credenciales quedan fuera del modelo exportado.

Esquema físico en SQLite, contratos concretos del puente Wails, detalles de distribución portable y ubicación de datos. Propuesta: núcleo común y tablas especializadas por medio dentro de una base, sin acoplar el formato de exportación al esquema interno. Ver [stack](tech-stack.md) y [backend](backend.md).
