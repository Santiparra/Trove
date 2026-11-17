# Desarrollo en Windows x64

## Herramientas

| Herramienta | Referencia verificada |
| --- | --- |
| Go | 1.27.4; el módulo requiere 1.26 o posterior |
| Node.js / npm | Node 26.13.0 y npm incluido |
| Wails CLI | 3.0.0 |
| Git | Versión estable |
| WebView2 Runtime | Necesario para ejecutar la ventana nativa |

Go: [descargas oficiales](https://go.dev/dl/). Node: [descargas oficiales](https://nodejs.org/). WebView2: [Microsoft](https://developer.microsoft.com/microsoft-edge/webview2/).

Después de instalar Go, desde cualquier carpeta:

```powershell
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0
go version
node --version
npm --version
wails3 doctor
```

PATH predeterminado: `C:\Program Files\Go\bin` y `%USERPROFILE%\go\bin`. Abrir una terminal nueva después de modificarlo; reiniciar también el editor si su terminal conserva el entorno anterior.

SQLite está integrado mediante un driver Go sin CGo. React, Vite, TypeScript y las herramientas de frontend son dependencias locales, no instalaciones globales.

## Preparación y ejecución

```powershell
go mod download
npm --prefix frontend ci
wails3 task dev
```

Vite escucha en `127.0.0.1:9245`. El frontend tiene recarga rápida y los cambios de Go reinician el backend. Detener con Ctrl+C. Los datos de desarrollo están separados de pruebas y uso habitual.

## Verificación

```powershell
powershell -NoProfile -File tools/verify.ps1
```

Ejecuta Go y frontend. `-CheckBindings` regenera el contrato Wails y exige que coincida con los archivos del commit; usarlo con los cambios de bindings ya incorporados a un commit. No es una prueba del navegador o ventana nativa completos.

```powershell
wails3 task bindings
npm --prefix frontend run format
wails3 task test
```

Las carpetas temporales del script quedan en `.local/`. Los tests usan datos sintéticos, no bibliotecas personales ni APIs reales.

## Compilar

Build rápida local de prueba:

```powershell
wails3 task build
.\bin\Trove-test.exe
```

Paquete identificable, con licencia, instrucciones, metadatos y checksums:

```powershell
powershell -NoProfile -File tools/build.ps1 -Version development -Profile test -BuildID local
powershell -NoProfile -File tools/verify-without-git.ps1
```

La salida está en `artifacts/portable/`. Para compilar no se necesita que la carpeta sea un repositorio Git. La revisión se inyecta explícitamente cuando está disponible; las compilaciones usan `-buildvcs=false`. El contrato sin Git reconstruye el ejecutable desde una copia de fuentes con los assets ya generados y prueba la identidad de ambos ejecutables desde directorios vacíos.

La versión de aplicación no es la versión del esquema SQLite ni la del respaldo JSON. Los scripts no incrementan archivos de versión ni crean commits.

## Integración

`dev` es la rama de integración de trabajo. Solo los PR cuyo destino es `main` y los pushes a `main` ejecutan GitHub Actions. Un push a `dev` por sí solo no inicia un workflow.

El check Windows valida contratos, frontend, bindings y paquete portable. El artefacto de PR usa perfil `test`. Después del merge, se repiten las verificaciones sobre el commit integrado y se publica el paquete `habitual` desde el artefacto de ese mismo run, sin recompilar en el job de publicación.

Ver [contratos](testing.md), [arquitectura](architecture.md) y [releases](releases/README.md).
