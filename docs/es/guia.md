# Cómo empezar y usar Trove

Trove es una biblioteca y bitácora multimedia local para Windows x64. Es un proyecto personal abierto a quien quiera usarlo o contribuir. La interfaz sigue en español; la [documentación pública principal](../../README.md) se mantiene en inglés.

Esta guía resume el uso disponible desde el corte **v0.1.7** de Libros y Novelas ligeras. Se revisa al preparar PRs a `main`: durante el desarrollo puede quedar por detrás de funciones todavía no publicadas. Las novedades de cada publicación están en el [historial de releases](../releases/README.md).

## 1. Descargar, ejecutar o desarrollar

### Descargar y ejecutar

1. Abrí [GitHub Releases](https://github.com/Santiparra/Trove/releases) y descargá el ZIP **Windows x64**.
2. Extraelo completo en una carpeta donde tengas permiso de escritura.
3. Abrí `Trove.exe`.

Necesitás [Microsoft WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/), normalmente disponible en Windows moderno. Para usar el portable no necesitás instalar Go, Node ni un servidor SQLite. No hay cuenta de Trove.

La biblioteca de uso habitual se guarda en `data/habitual/`, junto al ejecutable. Para actualizar, cerrá Trove, conservá esa carpeta y reemplazá los archivos de la aplicación por los de la nueva release. Copiar solo el EXE no transfiere la biblioteca; también podés empezar en otra carpeta y restaurar un respaldo.

### Desarrollo y contribuciones

Instalá Go, Node.js, Git, Wails CLI y WebView2 siguiendo las versiones y pasos de la [guía de desarrollo](../development.md). Después:

```powershell
git clone https://github.com/Santiparra/Trove.git
cd Trove
go mod download
npm --prefix frontend ci
wails3 task dev
```

El frontend se actualiza al editarlo y los cambios de Go reinician el backend. Detené el entorno con Ctrl+C. Desarrollo usa datos separados de los perfiles de prueba y uso habitual.

El trabajo cotidiano se integra en `dev`. Los PR a `main` ejecutan verificaciones Windows; integrar en `main` publica automáticamente si pasan. La documentación técnica, los contratos y los detalles de contribución están en inglés. La planificación y notas internas siguen en español.

## 2. Uso cotidiano

### Bibliotecas y registros

- **Libros** y **Novelas ligeras** tienen portales, búsquedas, series y operaciones independientes. En Novelas cada volumen se registra por separado; todavía no hay catálogo remoto para esa sección.
- Podés agregar manualmente un libro o volumen. Solo el título es obligatorio; autoría, serie, edición y fechas se pueden completar después.
- Desde una ficha podés cambiar estado, favorito y propiedad, guardar notas, editar o eliminar. Eliminar requiere confirmación.
- Los estados son **Por leer**, **A continuación**, **Leyendo**, **Leído**, **Pausado** y **Sin terminar**. Las fechas desconocidas pueden quedar vacías; marcar una lectura antigua como leída no inventa una finalización sin inicio.

### Buscar y agregar desde catálogo

1. En **Claves**, configurá una clave de lectura de Hardcover para el perfil. El catálogo es opcional; Open Library no necesita clave.
2. Usá la búsqueda principal por título, autor, serie o ISBN. Primero aparecen coincidencias locales y después resultados de Hardcover en el mismo flotante. **Ver todos los resultados** abre el listado.
3. Elegí un resultado para abrir su ficha. Ahí podés revisar información y descripción o **Elegir otra edición**. Pulsar el autor o la serie abre otra búsqueda.
4. **Agregar · Por leer** agrega con un clic. Elegir otro estado en su desplegable ejecuta el alta directamente. Una única serie se preselecciona y puede quitarse.

Las ediciones vinculadas que ya tenés abren su registro local. Las coincidencias que requieren revisión se presentan antes de vincular; no se fusionan registros automáticamente. **Refrescar metadata** conserva tus datos personales y, si falla, mantiene la información anterior.

### Proveedores y modo solo local

El indicador **Proveedores** muestra la salud de las APIs y permite **Comprobar ahora**. Los fallos y límites pueden pausar consultas temporalmente; una clave ausente o rechazada se distingue de una caída.

**Modo solo local** bloquea consultas y descargas remotas, cancela solicitudes en curso y se conserva al reiniciar el perfil. Podés seguir usando tu biblioteca, las descripciones guardadas, las portadas ya cacheadas y la gestión personal. Las portadas no descargadas muestran un marcador. Es útil para comprobar respaldos sin desconectar Internet.

### Respaldar y restaurar

1. Abrí **Respaldo y restauración** y elegí una carpeta distinta de los datos activos.
2. Trove crea una copia y después respalda los cambios automáticamente, conservando dos generaciones de SQLite y JSON. **Al día** confirma el resguardo local, no la subida a Drive/Dropbox.
3. Para recuperar, elegí **Actual** o **Anterior** por fecha y hora y pulsá **Revisar respaldo seleccionado**.
4. Revisá contenido y perfil, confirmá el reemplazo y pulsá **Restaurar y reemplazar biblioteca**. También podés elegir un archivo SQLite o JSON manualmente.

Restaurar reemplaza la biblioteca completa del perfil y conserva una copia previa; no mezcla registros ni importa datos de otros servicios. Cerrá otras instancias antes de hacerlo. Los respaldos no incluyen claves, portadas cacheadas ni preferencias del perfil. El modo solo local no cambia al restaurar.

Los éxitos se muestran como avisos flotantes sin mover la página. Los errores quedan junto a la operación. Para detalles, consultá la [guía completa](../usage.md), el [formato de respaldo](../backup-format.md) y las [novedades](../releases/README.md).

### Alcance actual

Las importaciones de servicios externos, los puntajes y las demás secciones multimedia todavía no están disponibles. Esta guía se amplía con recorridos nuevos al preparar cada PR de publicación, sin duplicar toda la documentación técnica ni mantener un segundo changelog.
