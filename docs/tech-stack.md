# Stack: opciones pendientes

El contenedor y los lenguajes no están elegidos. SQLite sí está aceptado. El usuario tiene algo de experiencia con Go y desea evaluar su uso en lugar de Rust. Se prefiere una experiencia portable con carpetas de salida configurables.

## Opciones de escritorio

| Opción | Lenguaje de núcleo | Ventajas para Trove | Costes / aspectos a comprobar |
| --- | --- | --- | --- |
| Electron + React | TypeScript/Node.js | Un lenguaje principal, ecosistema amplio, Chromium incluido y comportamiento visual más uniforme | Mayor tamaño y memoria; integración y empaquetado de SQLite según biblioteca |
| Wails + React | Go | Aprovecha experiencia previa, acceso nativo a archivos y SQLite, alternativa natural para núcleo Go | Dos lenguajes y frontera UI/Go; WebView2 en Windows; comprobar versión, bindings, distribución y driver SQLite |
| Tauri + React | Rust | Núcleo nativo y distribución generalmente menor que Electron | Aprendizaje de Rust; frontera UI/Rust; WebView2 en Windows |

Go no reemplaza directamente el backend Rust estándar de Tauri. Si se elige Go, Wails es la alternativa a evaluar; añadir un proceso Go a Tauri implicaría complejidad adicional que hoy no se justifica.

## Recomendación provisional

Para una app personal local y experiencia previa con Go, Wails merece compararse primero con Electron. Elegir Go por familiaridad y mantenibilidad, no porque Trove necesite concurrencia o rendimiento extraordinarios. Rust no es un requisito del producto.

Antes de aprobar stack: verificar documentación vigente de versiones, compatibilidad, distribución y acceso a SQLite. Esta comparación no es una validación técnica ejecutada.

## Componentes candidatos

- **Interfaz:** React, TypeScript estricto y Vite.
- **Persistencia confirmada:** SQLite embebido, con resguardo automático de datos propios y exportación abierta de fallback. Ver [persistencia y recuperación](backend.md).
- **Validación:** Zod para fronteras TypeScript; validación equivalente en Go/Rust cuando corresponda.
- **Acceso a datos:** Drizzle o SQL con migraciones si se elige Node; biblioteca/driver y migraciones propios del ecosistema si se elige Go o Rust.
- **Pruebas:** Vitest para lógica TypeScript, herramientas nativas para Go/Rust y recorridos de escritorio según soporte del contenedor. No asumir que Playwright automatiza igual las tres opciones.
- **Versionado:** Git; CI y estrategia de releases por definir.

No se necesita decidir ahora una librería de gráficos, componentes visuales ni todos los proveedores.
