# Trove

Biblioteca y bitácora personal multimedia para Windows. Centraliza el seguimiento de libros, anime y otros medios, con datos locales independientes de los servicios externos.

> **En planificación.** El repositorio contiene documentación de producto y arquitectura. Todavía no hay una aplicación ejecutable, comandos de arranque ni releases. Trove es un nombre provisional.

## Objetivo

- Gestionar una biblioteca personal sin cuenta de Trove ni conexión obligatoria.
- Registrar estados, puntajes, notas, progreso y ciclos de lectura o visualización.
- Crear obras y ediciones manuales y vincular metadata externa de forma opcional.
- Importar archivos exportados desde servicios especializados.
- Conservar datos mediante respaldos automáticos y formatos abiertos de recuperación.

El primer medio será **libros**, con **The StoryGraph** como primera fuente de importación. Las funcionalidades descritas son objetivos de desarrollo, no características implementadas.

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

En esta etapa, clonar permite consultar y trabajar sobre la documentación. Las instrucciones de instalación de dependencias del proyecto y ejecución se incorporarán cuando exista la aplicación.

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

React, TypeScript, Vite y el driver SQLite serán dependencias versionadas del proyecto; no requieren instalación global. SQLite se integra en la aplicación y no necesita un servidor separado.

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
