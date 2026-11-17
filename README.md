# Trove

Biblioteca y bitácora multimedia local para **Windows x64**. Los datos personales pertenecen a la biblioteca local; los proveedores externos no son necesarios para registrar libros.

## Funciones disponibles

- Biblioteca manual de libros con autoría, serie y orden, edición, estado y fechas.
- Portal de lectura, búsqueda local y navegación por series.
- Notas, favoritos, propiedad y cambios rápidos de estado.
- Respaldo automático SQLite + JSON y restauración completa con validación y copia previa.
- Perfiles separados de desarrollo, prueba y uso habitual.

Las demás secciones multimedia están en construcción. No están implementadas las importaciones de servicios externos, proveedores de metadata, portadas reales ni puntajes.

## Descargar y ejecutar

Las entregas se publican en [Releases](https://github.com/Santiparra/Trove/releases) cuando un cambio integrado en `main` supera las verificaciones. Mientras no haya una publicación disponible se puede compilar desde el código.

1. Descargar el ZIP **Windows x64** de una release.
2. Extraerlo en una carpeta con permiso de escritura.
3. Abrir `Trove.exe`.

Requisito de ejecución: [Microsoft WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/), normalmente disponible en Windows moderno. No requiere Go, Node, un servidor SQLite ni cuenta de Trove. El paquete portable no está firmado digitalmente.

Las releases usan `data/habitual/` junto al ejecutable. Los artefactos de PR usan un perfil de prueba independiente. Copiar solo el EXE no transfiere una biblioteca: conservar los datos o restaurar un respaldo.

## Datos y recuperación

Configurar una carpeta desde **Respaldo y restauración**. Trove mantiene dos generaciones completas de SQLite y JSON; un cliente de Drive/Dropbox puede sincronizar esa carpeta. El indicador confirma resguardo local, no subida a la nube.

El JSON documentado conserva identidades y relaciones sin depender de SQLite. Restaurar reemplaza el perfil seleccionado, no mezcla bibliotecas ni importa datos de otros servicios. Los archivos no tienen cifrado propio.

## Desarrollo y contribuciones

```powershell
git clone https://github.com/Santiparra/Trove.git
cd Trove
go mod download
npm --prefix frontend ci
wails3 task dev
```

Antes de ejecutar esos comandos, instalar las herramientas indicadas en [desarrollo](docs/development.md). El código de trabajo se integra en `dev`; los PR destinados a `main` ejecutan las verificaciones Windows. Integrar en `main` inicia la publicación automática si las verificaciones pasan.

## Documentación

| Tema | Referencia |
| --- | --- |
| Uso y almacenamiento | [Guía de uso](docs/usage.md) |
| Entorno y comandos | [Desarrollo](docs/development.md) |
| Responsabilidades y límites | [Arquitectura](docs/architecture.md) |
| Recuperación portable | [Formato de respaldo](docs/backup-format.md) |
| Criterios verificables | [Contratos y pruebas](docs/testing.md) |
| Publicación y artefactos | [Releases](docs/releases/README.md) |
| Alcance de privacidad | [Seguridad](docs/security.md) |

## Licencia

[MIT](LICENSE).
