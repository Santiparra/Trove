# Entorno de desarrollo recomendado para Windows

Estado: stack Wails v3 + Go + React/TypeScript + Vite + SQLite confirmado para Windows x64. Herramientas y diagnóstico comprobados; todavía no hay una aplicación ni dependencias de proyecto instaladas.

## Herramientas a instalar

1. **Go:** rama 1.26 con parche vigente, siguiendo la recomendación de Wails v3, o una estable posterior compatible. La guía consultada exige Go 1.26 o posterior.
2. **Node.js LTS y npm:** para construir React, TypeScript y Vite. Node 26 figura como LTS en la página oficial consultada; no usar la rama alpha. Una instalación LTS existente compatible con el frontend podría servir sin reinstalar.
3. **WebView2 Runtime:** comprobar disponibilidad en Windows; instalar desde Microsoft si falta.
4. **Wails CLI v3.0.0:** versión estable identificada en la documentación oficial consultada. Instalar después de Go y de asegurar su directorio de binarios en PATH.
5. **Git:** ya utilizado por el proyecto; no requiere reinstalación si funciona.

Comando de instalación del CLI:

```powershell
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0
```

El directorio por defecto del CLI es `%USERPROFILE%\go\bin`, salvo configuración distinta de Go. Si no se reconoce `wails3`, revisar PATH y abrir una terminal nueva.

## Verificación del entorno, sin crear proyecto

```powershell
go version
node --version
npm --version
git --version
wails3 doctor
```

El asistente `wails3 setup` figura como experimental; se recomienda instalación manual y diagnóstico con `doctor` para esta preparación.

## Dependencias que no se instalan globalmente

React, TypeScript y Vite serán dependencias versionadas del proyecto cuando se autorice implementar. SQLite se integrará mediante un driver Go, no como servidor o programa separado.

Propuesta de driver: `modernc.org/sqlite`, compatible con `database/sql` y sin CGo. Evita requerir compilador C por SQLite en este escenario; no se agregó al proyecto ni se fijó todavía versión. Sus costes de rendimiento y compatibilidad se evaluarán con el recorrido real; no equivale a garantizar el mismo rendimiento que SQLite compilado en C.

Con Wails v3 en Windows y ese driver, la guía oficial no exige Visual Studio C++ Build Tools, GCC ni Rust como prerrequisitos básicos. Dependencias futuras podrían cambiarlo. No se necesita Docker, una base de datos servidor ni NSIS para comenzar; NSIS corresponde a ciertos instaladores, no al objetivo inicial portable.

## Uso después de formatear

Para ejecutar un paquete compilado: Trove, WebView2 disponible y datos/restauración. Go, Node/npm y Wails CLI son herramientas de desarrollo, no requisitos del usuario final. La forma exacta del paquete portable debe comprobarse al implementar y preparar una release.

## Fuentes consultadas el 2026-11-15

- [Wails v3: instalación](https://v3.wails.io/quick-start/installation/): v3.0.0 estable, Go 1.26+, WebView2 y CLI.
- [Go: descargas](https://go.dev/dl/).
- [Node: ramas y estado LTS](https://nodejs.org/en/about/previous-releases).
- [WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/).
- [Driver modernc SQLite](https://pkg.go.dev/modernc.org/sqlite).
- [Tauri: prerrequisitos](https://v2.tauri.app/start/prerequisites/): comparación con Rust y herramientas C++ de Windows.

## Comprobación local del 2026-11-15

Comandos de versión y `wails3 doctor` ejecutados desde el repo:

| Herramienta | Resultado |
| --- | --- |
| Go | go1.27.4 windows/amd64 |
| Node.js | v26.13.0 |
| npm | 11.20.0 |
| Git | 2.56.0.windows.4 |
| Wails CLI | v3.0.0 |
| WebView2 | 157.0.4341.46 |

`doctor` terminó con «No issues found» y «Your system is ready for Wails development!». También mostró `CGO_ENABLED=0`, coherente con la propuesta de driver SQLite sin CGo.

No están instalados Android SDK, Docker, NSIS, MSIX Packaging Tool, MakeAppx ni SignTool; no bloquean el desarrollo inicial para Windows portable. Se evaluarían solo si se solicita una distribución que los necesite. La firma de ejecutables tampoco está configurada.

Alcance comprobado: herramientas accesibles en PATH y prerrequisitos detectados por Wails. React/Vite/TypeScript y el driver SQLite todavía no están incorporados al repo, por lo que no se verificó una compilación integrada, ejecución de Trove, persistencia ni empaquetado. No se realizaron instalaciones ni benchmarks.
