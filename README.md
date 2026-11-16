# Trove

Biblioteca y bitácora personal multimedia para Windows. Centraliza el seguimiento de libros, anime y otros medios, con datos locales independientes de los servicios externos.

> **Primera entrega local en desarrollo.** Navegación por medios y biblioteca manual de libros con SQLite. Las otras secciones están en construcción. No hay una release publicada ni versión pública asignada. Trove es un nombre provisional.

## Objetivo

- Gestionar una biblioteca personal sin cuenta de Trove ni conexión obligatoria.
- Registrar estados, puntajes, notas, progreso y ciclos de lectura o visualización.
- Crear obras y ediciones manuales y vincular metadata externa de forma opcional.
- Importar archivos exportados desde servicios especializados.
- Conservar datos mediante respaldos automáticos y formatos abiertos de recuperación.

El primer medio es **libros**. El portal muestra estados y cantidades, lecturas actuales, últimos leídos con fecha registrada y distribución visual. Desde allí se accede a los listados, la búsqueda y fichas completas para crear, editar, cambiar rápidamente el estado o eliminar un libro con confirmación. Se conservan autoría, Serie y número de orden, edición mínima, notas y fechas. Las importaciones, proveedores, portadas reales, puntajes, relecturas avanzadas y respaldos automáticos todavía no están implementados. The StoryGraph será la primera fuente de importación.

## Stack

| Área | Tecnología |
| --- | --- |
| Escritorio | Wails v3 |
| Núcleo y lógica de negocio | Go |
| Interfaz | React + TypeScript + Vite |
| Persistencia local | SQLite |
| Plataforma inicial | Windows x64 |

Interfaz y documentación en español; nombres de código en inglés. Los títulos de las obras conservan su idioma, independientemente del idioma de la interfaz.

## Desarrollo

### Obtener el repositorio

```powershell
git clone https://github.com/Santiparra/Trove.git
cd Trove
```

### Instalar dependencias del proyecto

Desde la raíz, después de instalar las herramientas de la sección siguiente:

```powershell
go mod download
npm --prefix frontend ci
```

No es necesario instalar SQLite: se integra mediante `modernc.org/sqlite`, sin CGo. No se requiere un compilador C por este driver.

### Ejecutar en desarrollo

```powershell
wails3 task dev
```

Genera bindings, compila el backend y abre la ventana Wails con frontend Vite en `127.0.0.1:9245`. Los cambios del frontend se recargan; los de Go reinician el backend. Detener con `Ctrl+C` en esa terminal. El puerto debe estar disponible.

### Comprobar y compilar una build de prueba

```powershell
wails3 task test
wails3 task build
.\bin\Trove-test.exe
```

`test` ejecuta las pruebas Go y frontend. `build` compila el frontend y el ejecutable con assets embebidos, marcado como **PRUEBA / no publicado**, sin asignar versión de release. No requiere dejar un servidor de desarrollo activo para usarlo. WebView2 Runtime debe estar disponible en Windows.

La build está en `bin/Trove-test.exe`. Se puede copiar a otra carpeta y ejecutar; sus datos nuevos quedan junto al ejecutable bajo `data/test/`. Copiar solo el EXE no copia una biblioteca existente: para conservar datos de esta prueba, trasladar también esa carpeta con la app cerrada. La build no está firmada.

### Datos separados

| Perfil | Ubicación predeterminada |
| --- | --- |
| Desarrollo | `.local/development/` en el directorio desde el que se inicia |
| Build de prueba | `data/test/` junto al ejecutable |
| Habitual (solo selección explícita) | `data/habitual/` junto al ejecutable |

El perfil y la ruta activa se pueden consultar en la etiqueta superior de la app. Esta entrega debe probarse con datos reemplazables: aún no hay backup/restauración automáticos.

Opcionalmente se puede elegir una raíz de datos; el nombre del perfil siempre se añade para mantener aislamiento:

```powershell
.\bin\Trove-test.exe --profile test --data-dir "C:\TrovePruebas"
```

El ejemplo usa `C:\TrovePruebas\test\`. La base `trove.sqlite` y la configuración WebView que gestione el runtime no son material para publicar en Git. La app crea el esquema al primer arranque y rechaza esquemas futuros que no pueda interpretar.

### Recorrido manual de aceptación

1. Recorrer las secciones; solo Libros está funcional.
2. Crear un libro sintético, con autoría, Serie y número, sin ISBN ni proveedor.
3. Crear otro libro de la misma serie usando la sugerencia del formulario y otro número.
4. Buscar por título, editar estado, notas, edición y fechas conocidas; dejar las desconocidas vacías.
5. Cerrar completamente y reabrir. Verificar datos y vínculos en el perfil de prueba.
6. Confirmar que desarrollo y build de prueba no muestran la misma biblioteca por defecto.
7. Comprobar cantidades y distribución del portal, entrar por estado y cambiarlo desde el selector rápido. Verificar que no se sobrescriben notas ni se fabrican fechas.
8. Añadir fecha de finalización a una lectura terminada y comprobar su aparición entre los últimos leídos. Las portadas son marcadores locales provisionales.
9. Abrir una ficha y su formulario a página completa. Cancelar la eliminación y luego probarla con un libro sintético prescindible; comprobar que otro libro de su serie se conserva.

Las pruebas automáticas verifican persistencia/reapertura y errores de guardado. No sustituyen la validación visual del paquete; el arranque desde una instalación limpia y la interacción manual deben comprobarse antes de considerar el candidato aprobado.

### Requisitos para Windows

| Herramienta | Requisito |
| --- | --- |
| [Go](https://go.dev/dl/) | 1.26 o posterior compatible con Wails v3 |
| [Node.js y npm](https://nodejs.org/) | Rama LTS compatible; npm incluido con Node |
| [Git](https://git-scm.com/downloads) | Versión estable |
| [WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/) | Disponible en el sistema; instalar si falta |
| Wails CLI | v3.0.0 como referencia inicial |

Instalar el CLI después de Go. El comando se puede ejecutar desde cualquier directorio y no genera un proyecto:

```powershell
go install github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0
```

Comprobar las herramientas y prerrequisitos:

```powershell
go version
node --version
npm --version
git --version
wails3 doctor
```

React, TypeScript, Vite y el driver SQLite son dependencias versionadas del proyecto; no requieren instalación global. Los lockfiles de npm y Go deben conservarse.

### Resolver problemas de PATH

En instalaciones predeterminadas de Windows, el PATH debe incluir:

- `C:\Program Files\Go\bin`, para el ejecutable de Go.
- `%USERPROFILE%\go\bin`, para herramientas como `wails3.exe`.

Las rutas pueden variar si se personalizaron la instalación, `GOBIN` o `GOPATH`. Después de modificar el entorno, abrir una terminal nueva; en terminales integradas puede ser necesario reiniciar también el editor o la aplicación anfitriona.

Para diagnosticar si el CLI está instalado pero no se encuentra en PATH:

```powershell
& "$env:USERPROFILE\go\bin\wails3.exe" doctor
```

La [guía del entorno](docs/development-environment.md) detalla las dependencias, fuentes y alcance de las comprobaciones realizadas. Las herramientas de compilación no serán necesarias para ejecutar una release; los requisitos de distribución se documentarán al preparar el paquete.

## Documentación

| Tema | Documento |
| --- | --- |
| Producto | [Visión y alcance](docs/vision.md) · [Decisiones](docs/decisions.md) · [Roadmap](docs/roadmap.md) |
| Diseño técnico | [Arquitectura](docs/architecture.md) · [Dominio](docs/domain.md) · [Stack](docs/tech-stack.md) |
| Componentes | [Frontend](docs/frontend.md) · [Backend y persistencia](docs/backend.md) · [Integraciones](docs/integrations.md) |
| Importaciones | [Análisis de exportaciones de ejemplo](docs/import-analysis.md) |
| Entorno | [Preparación de desarrollo](docs/development-environment.md) |
| Calidad | [Estrategia de pruebas](docs/testing.md) |
| Operación | [Seguridad](docs/security.md) · [Durabilidad](docs/longevity.md) · [Releases](docs/releases/README.md) |

La documentación distingue decisiones confirmadas, propuestas y asuntos pendientes. Las exportaciones personales usadas para analizar formatos no forman parte de los datos públicos del repositorio.

## Licencia

Consultar [LICENSE](LICENSE).
