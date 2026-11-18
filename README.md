# Trove

Biblioteca y bitácora multimedia local para **Windows x64**. Los datos personales pertenecen a la biblioteca local; los proveedores externos no son necesarios para registrar libros.

**Novedades de este corte:** catálogo de Libros, biblioteca manual de Novelas ligeras, modo solo local y mejoras de búsqueda, fichas y restauración. Ver [qué cambia respecto a v0.1.4](docs/releases/catalogo-libros-y-novelas.md) y el [historial de publicaciones](docs/releases/README.md#historial-de-novedades).

## Funciones disponibles

- **Libros:** búsqueda unificada en la biblioteca y Hardcover, con Open Library como complemento seguro por ISBN de la edición elegida.
- Ficha compartida para resultados locales y de catálogo, con portadas, descripción, selección de ediciones y navegación por autoría y series.
- Alta manual o desde catálogo, vinculación con registros existentes y metadata guardada sin reemplazar notas ni seguimiento personal.
- **Novelas ligeras:** biblioteca manual independiente, con autoría, series y orden, edición, estado y fechas. Cada volumen es una entrada; todavía no tiene proveedor remoto.
- Portal de lectura, notas, favoritos, propiedad y gestión rápida de estados. Elegir un estado al agregar ejecuta el alta directamente.
- Indicador de salud de proveedores, diagnóstico manual y pausas automáticas ante fallos o límites.
- **Modo solo local** persistente por perfil: bloquea consultas y descargas, conservando biblioteca, metadata y portadas ya cacheadas.
- Avisos de éxito flotantes que no desplazan el contenido.
- Respaldo automático SQLite + JSON y restauración **Actual / Anterior** con fecha, hora, validación y copia previa al reemplazo.
- Perfiles separados de desarrollo, prueba y uso habitual.

El catálogo es opcional. Hardcover requiere configurar una clave de lectura; Open Library no requiere clave. El alta manual, la gestión personal y la lectura de los datos guardados siguen disponibles sin conexión. Ver [uso y configuración](docs/usage.md).

Las demás secciones multimedia están en construcción. No están implementadas las importaciones de servicios externos ni puntajes. Listados, búsquedas, series y operaciones se acotan al medio seleccionado.

## Descargar y ejecutar

Las entregas se publican en [Releases](https://github.com/Santiparra/Trove/releases) cuando un cambio integrado en `main` supera las verificaciones. Mientras no haya una publicación disponible se puede compilar desde el código.

1. Descargar el ZIP **Windows x64** de una release.
2. Extraerlo en una carpeta con permiso de escritura.
3. Abrir `Trove.exe`.

Para actualizar, cerrar Trove y conservar la carpeta `data/` del portable anterior; extraer la nueva entrega en esa ubicación reemplazando los archivos de la aplicación. También se puede usar una carpeta nueva y restaurar un respaldo. Las notas de cada corte describen sus cambios y límites.

Requisito de ejecución: [Microsoft WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/), normalmente disponible en Windows moderno. No requiere Go, Node, un servidor SQLite ni cuenta de Trove. El paquete portable no está firmado digitalmente.

Las releases usan `data/habitual/` junto al ejecutable. Los artefactos de PR usan un perfil de prueba independiente. Copiar solo el EXE no transfiere una biblioteca: conservar los datos o restaurar un respaldo.

## Datos y recuperación

Configurar una carpeta desde **Respaldo y restauración**. Trove mantiene dos generaciones completas de SQLite y JSON; un cliente de Drive/Dropbox puede sincronizar esa carpeta. El indicador confirma resguardo local, no subida a la nube.

El selector muestra **Actual** y **Anterior** con fecha y hora. También permite elegir un archivo de respaldo manualmente. Las claves, portadas cacheadas y preferencias del perfil, incluido el modo solo local, quedan fuera de la exportación.

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
