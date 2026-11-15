# Stack tecnológico

Stack confirmado: **Wails v3 + Go + React/TypeScript + Vite + SQLite**, para Windows x64. Se prefiere una experiencia portable con carpetas de salida configurables. Driver concreto, bibliotecas auxiliares y detalles de distribución pendientes de selección técnica.

## Alternativas evaluadas

| Opción | Lenguaje de núcleo | Ventajas para Trove | Costes / aspectos a comprobar |
| --- | --- | --- | --- |
| Electron + React | TypeScript/Node.js | Un lenguaje principal, ecosistema amplio, Chromium incluido y comportamiento visual más uniforme | Mayor tamaño y memoria; integración y empaquetado de SQLite según biblioteca |
| Wails + React | Go | Aprovecha experiencia previa, acceso nativo a archivos y SQLite, alternativa natural para núcleo Go | Dos lenguajes y frontera UI/Go; WebView2 en Windows; comprobar versión, bindings, distribución y driver SQLite |
| Tauri + React | Rust | Núcleo nativo y distribución generalmente menor que Electron | Aprendizaje de Rust; frontera UI/Rust; WebView2 en Windows |

Go no reemplaza directamente el backend Rust estándar de Tauri. Si se elige Go, Wails es la alternativa a evaluar; añadir un proceso Go a Tauri implicaría complejidad adicional que hoy no se justifica.

## Motivos de la elección

Wails permite aprovechar experiencia previa con Go en una aplicación personal local. La elección prioriza familiaridad y mantenibilidad; Trove no requiere concurrencia ni rendimiento extraordinarios. Rust no es un requisito del producto.

Stack aprobado y herramientas verificadas mediante versiones y `wails3 doctor`; aún no existe una compilación integrada de Trove. Distribución portable y acceso a SQLite se comprobarán al implementar sus recorridos.

Revisión documental del 2026-11-15: la documentación oficial presenta Wails v3.0.0 como estable y exige Go 1.26+. Propuesta concreta: Wails v3 + Go + React/TypeScript + Vite + SQLite mediante `database/sql` y un driver sin CGo como `modernc.org/sqlite`; npm para el frontend. Ver [preparación del entorno](development-environment.md). La disponibilidad de documentación no sustituye una comprobación del paquete portable ni un build real.

## Wails/Go frente a Tauri/Rust para Trove

Ambos utilizan WebView2 en Windows: cambiar el lenguaje del núcleo no elimina el coste de renderizar la misma interfaz. Go tiene runtime y recolector de basura; Rust permite control más fino de memoria sin GC. Para biblioteca local, importaciones y consultas a proveedores, no hay evidencia de que esa diferencia justifique aprender Rust; red, consultas e interfaz pueden dominar la latencia.

Tauri ofrece un modelo explícito de permisos/capabilities y un catálogo de plugins que merece considerar si se necesitan esas capacidades. Wails encaja bien con servicios escritos en Go y aprovecha la experiencia del usuario; los límites de red, acceso a archivos y operaciones expuestas deben diseñarse igualmente. Ninguno garantiza durabilidad por diez años. Mantener el dominio y adaptadores independientes del contenedor reduce el coste de una eventual migración.

Recomendación: Wails/Go para Trove, sin prometer cifras de memoria, arranque o superioridad global de rendimiento. No se hicieron benchmarks propios.

## Componentes candidatos

- **Interfaz:** React, TypeScript estricto y Vite.
- **Persistencia confirmada:** SQLite embebido, con resguardo automático de datos propios y exportación abierta de fallback. Ver [persistencia y recuperación](backend.md).
- **Validación:** Zod para fronteras TypeScript; validación equivalente en Go/Rust cuando corresponda.
- **Acceso a datos:** Drizzle o SQL con migraciones si se elige Node; biblioteca/driver y migraciones propios del ecosistema si se elige Go o Rust.
- **Pruebas:** Vitest para lógica TypeScript, herramientas nativas para Go/Rust y recorridos de escritorio según soporte del contenedor. No asumir que Playwright automatiza igual las tres opciones.
- **Versionado:** Git; CI y estrategia de releases por definir.

No se necesita decidir ahora una librería de gráficos, componentes visuales ni todos los proveedores.

## Desarrollar frente a ejecutar una release

Para desarrollar/compilar se necesitan toolchains, dependencias y herramientas del stack elegido. Para usar una release ya compilada no se necesitan esos compiladores ni instalar un servidor SQLite.

| Opción | Entorno de desarrollo orientativo | Ejecución de release |
| --- | --- | --- |
| Wails/Go | Go, Node y gestor de paquetes para frontend, herramientas Wails; herramientas nativas si lo requiere el driver SQLite/build | Ejecutable/paquete de Trove con frontend y SQLite integrado según driver; WebView2 disponible o distribuido según estrategia |
| Electron/TypeScript | Node, gestor de paquetes y herramientas de empaquetado; toolchain nativa si lo requiere SQLite | Paquete portable con ejecutable, Chromium, Node y recursos necesarios; puede ser una carpeta o un autoextraíble, no necesariamente un único EXE autónomo |
| Tauri/Rust | Rust, Node para frontend y herramientas de compilación de Windows requeridas por la versión | Ejecutable/paquete con frontend y SQLite integrado según biblioteca; WebView2 disponible o distribuido según estrategia |

Después de formatear para **usar** Trove: restaurar paquete, datos y configuración; instalar/proveer solo requisitos de ejecución que no estén incluidos. Para **seguir desarrollando**: reconstruir también el entorno de desarrollo.

No hay una lista definitiva de instalaciones hasta elegir versiones, driver SQLite y empaquetado. En Windows, WebView2 puede estar ya instalado, pero no se debe prometer portabilidad completa sin comprobar una instalación limpia. Un runtime WebView2 de versión fija distribuido junto a la app es una alternativa a evaluar por tamaño, soporte del contenedor y actualizaciones de seguridad.
