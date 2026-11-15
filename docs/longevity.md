# Durabilidad y mantenimiento

## Objetivo expresado por el usuario

Poder terminar el proyecto y seguir utilizando la aplicación durante años, sin mantenimiento obligatorio por cada versión nueva de sus herramientas. Es un objetivo de diseño, no una garantía de compatibilidad por diez años.

## Qué cambia y qué no

- **Uso local:** un binario ya compilado no necesita recompilarse porque aparezca una versión nueva de Go, Node, Rust, React o una biblioteca. Los datos personales y snapshots guardados deben seguir disponibles sin Internet.
- **Sistema y runtime:** cambios de Windows, arquitectura o dependencias de ejecución pueden exigir adaptación. Wails/Tauri dependen de WebView2; Electron incluye su runtime web, con otros costes de tamaño y mantenimiento.
- **Proveedores:** APIs, formatos, autenticación, certificados, cuotas o condiciones pueden cambiar aunque el servicio siga existiendo. El adaptador afectado podría necesitar actualización; no debe bloquear la biblioteca local.
- **Seguridad:** una versión congelada también conserva vulnerabilidades. No exigir actualización para abrir la biblioteca no implica que seguir consultando Internet con software antiguo esté libre de riesgo.
- **Restauración:** conservar identificadores permite intentar recuperar metadata, no garantiza disponibilidad futura. Los datos textuales respaldados reducen esa dependencia.

## Recomendaciones pendientes de aprobación

- Ninguna activación ni comprobación online obligatoria para abrir o editar datos propios.
- Actualizaciones elegidas por el usuario, sin caducidad artificial ni bloqueo por versión antigua.
- Separar fallos de integración del funcionamiento local y permitir desactivar un proveedor.
- Versionar esquema y formatos de exportación; hacer backup antes de migraciones destructivas y documentar compatibilidad. No asumir que una versión vieja puede abrir una base migrada por una nueva.
- Conservar junto al proyecto las versiones exactas de dependencias y las instrucciones de compilación; guardar los artefactos publicados. Esto ayuda a reproducir builds, pero no garantiza que herramientas y servicios externos sigan disponibles indefinidamente.

El compromiso razonable es maximizar la continuidad local y minimizar las causas de mantenimiento forzoso; no prometer ausencia total de mantenimiento.
