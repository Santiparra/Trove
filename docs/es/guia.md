# Cómo empezar y usar Trove

Trove es una biblioteca y bitácora multimedia local para Windows x64. Es un proyecto personal abierto a quien quiera usarlo o contribuir. La interfaz sigue en español; la [documentación pública principal](../../README.md) se mantiene en inglés.

Guía revisada para el corte de Novelas enriquecidas posterior a v0.1.7. Se actualiza al preparar PRs a `main`, no con cada iteración de desarrollo. El número de publicación lo asigna Actions; las novedades están en el [historial de releases](../releases/README.md).

## 1. Descargar, ejecutar o desarrollar

### Descargar y ejecutar

1. Abrí [GitHub Releases](https://github.com/Santiparra/Trove/releases) y descargá el ZIP **Windows x64**.
2. Extraelo completo en una carpeta donde tengas permiso de escritura.
3. Abrí `Trove.exe`.

Necesitás [Microsoft WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/), normalmente disponible en Windows moderno. Para usar el portable no necesitás instalar Go, Node ni un servidor SQLite. No hay cuenta de Trove.

La biblioteca habitual se guarda en `data/habitual/`, junto al ejecutable. Para actualizar, cerrá Trove, conservá esa carpeta y reemplazá los archivos de la aplicación. Copiar solo el EXE no transfiere una biblioteca; también podés empezar en otra carpeta y restaurar un respaldo.

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

El trabajo cotidiano se integra en `dev`. Los PR a `main` ejecutan verificaciones Windows; integrar en `main` publica automáticamente si pasan. La documentación técnica publicada está en inglés. Conversaciones, planificación y notas internas siguen en español. Al preparar cada PR se consolida la documentación y se revisa esta guía breve.

## 2. Uso cotidiano

### Libros

- Podés agregar manualmente o buscar por título, autor, serie o ISBN. La búsqueda reúne coincidencias locales y de Hardcover en el mismo flotante; **Ver todos los resultados** abre el listado.
- En **Claves**, **Obtener clave en Hardcover** abre la página oficial para conseguirla. El indicador amarillo avisa cuando falta configurar el acceso.
- La ficha reúne portada, descripción, edición y gestión. Autor y serie son clicables. **Agregar · Por leer** agrega con un clic; elegir otro estado ejecuta el alta directamente.
- La vinculación y el refresco conservan tus datos personales. Hardcover es el principal; Open Library complementa la edición por ISBN seguro.

### Novelas ligeras

- **Añadir novela** crea la entrada principal; cada volumen queda dentro de su ficha. Libros y Novelas mantienen datos y operaciones separados.
- La búsqueda local y de **RanobeDB** está unificada. No hace falta escribir `(Light Novel)`: la desambiguación se realiza detrás. **Hardcover** solo complementa publicaciones verificadas por ISBN; Open Library no interviene en Novelas.
- **Elegir volumen o publicación** permite revisar idioma y formato. Los registros ya vinculados abren su copia local; las coincidencias ambiguas se revisan antes de vincular.
- La galería muestra portadas de volúmenes registrados y futuros disponibles. El círculo indica el estado y se resalta el siguiente pendiente. Navegar no registra ni empieza una lectura.
- El portal reúne **Tu biblioteca**, **Leyendo ahora**, **A continuación** y **Últimos volúmenes leídos**. Estos últimos requieren fecha conocida. Los listados muestran miniaturas y **Todas** incluye todos los estados.

### Progreso y gestión

- El progreso cuenta volúmenes terminados, no la numeración impresa. Si leíste 1–6 y 6.5, son **7 completados** y el siguiente puede llamarse **7**. Las partes y etiquetas se muestran por separado.
- En **Editar novela y progreso**, **Voy por el volumen** identifica un libro del catálogo; no calcula X−1 desde una etiqueta decimal. Podés informar un total o dejarlo desconocido.
- **Empezar volumen** pasa a Leyendo y completa el inicio si falta. **Volumen terminado** avanza a A continuación sin iniciar el siguiente. Al llegar a un total conocido se ofrece completar la novela sin crear N+1.
- Las flechas del contador ajustan progreso sin inventar ni cambiar fechas de lectura. Se admiten cantidades desconocidas y novelas Leídas como `?/?` o `n/?`.
- Cada volumen conserva título opcional, datos de edición, sinopsis, notas, propiedad, favorito y fechas. Las notas de novela y volumen son independientes.
- **Identificadores y fuentes**, antes de la galería, contiene origen de los datos y acciones de catálogo. **Refrescar catálogo** conserva el último snapshot válido si falla. Las URLs de las descripciones se pueden abrir o copiar.
- Una advertencia de publicación preferida indica que falta una coincidencia verificable de idioma/formato. No equivale a falta de ISBN: sin ISBN puede funcionar RanobeDB, pero no el complemento Hardcover.
- Los perfiles cargan por etapas con skeletons y consultas espaciadas; no se descargan los detalles de todos los volúmenes al abrir una novela.

Los estados son **Por leer**, **A continuación**, **Leyendo**, **Leído/Leída**, **Pausado** y **Sin terminar**. Las fechas desconocidas pueden quedar vacías. Eliminar requiere confirmación; borrar una novela incluye sus volúmenes, sin afectar otros registros.

### Proveedores y modo solo local

**Proveedores** muestra salud, tiempos y pausas de las APIs y permite **Comprobar ahora**. Las claves ausentes o rechazadas se distinguen de una caída. No hay sondeos remotos continuos.

**Modo solo local** bloquea consultas y descargas remotas, cancela las que estén en curso y se conserva al reiniciar el perfil. Siguen disponibles la biblioteca, los snapshots y las portadas ya cacheadas. Es útil para comprobar respaldos sin desconectar Internet.

### Respaldar y restaurar

1. Abrí **Respaldo y restauración** y elegí una carpeta distinta de los datos activos.
2. Trove respalda automáticamente y conserva dos generaciones de SQLite y JSON. **Al día** confirma archivos locales, no su subida a Drive/Dropbox.
3. Elegí **Actual** o **Anterior** por fecha y hora y pulsá **Revisar respaldo seleccionado**. También podés elegir un archivo manualmente.
4. Revisá contenido y perfil, confirmá y pulsá **Restaurar y reemplazar biblioteca**.

Restaurar reemplaza toda la biblioteca del perfil y conserva una copia previa; no mezcla registros ni importa otros servicios. Cerrá otras instancias antes. Los respaldos no incluyen claves, portadas cacheadas ni preferencias. El modo solo local no cambia al restaurar.

**Datos del modelo anterior:** los volúmenes independientes de v0.1.7 no se convierten ni borran; permanecen almacenados y respaldados, pero no aparecen como novelas agrupadas en la nueva vista. Libros se conserva. Las versiones anteriores pueden no entender los nuevos datos o respaldos.

Los éxitos aparecen como avisos flotantes; los errores quedan junto a la operación. Para detalles, consultá la [guía completa](../usage.md), el [formato de respaldo](../backup-format.md) y las [novedades](../releases/README.md).

### Alcance actual

Las importaciones, los puntajes, los filtros por clasificación y la gestión avanzada de copias/relecturas todavía no están disponibles. Esta guía se amplía al preparar cada PR de publicación, sin duplicar la documentación técnica ni mantener un segundo changelog.
