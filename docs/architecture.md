# Arquitectura

Trove es una aplicación modular local: interfaz React/TypeScript en Wails v3, casos de uso Go y SQLite embebido. No requiere servidor de base de datos ni cuenta propia.

## Responsabilidades

| Componente | Responsabilidad |
| --- | --- |
| `frontend/src/books` | Presentación, formularios, navegación y estado de interacción |
| `frontend/src/backup` | Configuración visible, estado y confirmación de recuperación |
| `frontend/src/catalog` | Consulta, revisión de edición, vinculación y presentación de portadas locales |
| `internal/desktop` | Adaptadores Wails, contratos de entrada/salida y diálogos nativos |
| `internal/books` | Identidad, validaciones y casos de uso de biblioteca |
| `internal/storage/sqlite` | SQL, transacciones, migraciones y adaptación de snapshots |
| `internal/backup` | Formato lógico, validación, generaciones y restauración mediante un contrato de backend |
| `internal/config` | Resolución de perfiles y ubicación de datos |
| `internal/catalog` | Contratos bibliográficos, ISBN, cascada y procedencia independientes de proveedores |
| `internal/providers` | Adaptadores Hardcover/Open Library y caché local de portadas |
| `internal/credentials` | Protección DPAPI de la clave de catálogo por perfil |
| `main.go` | Composición de dependencias y ciclo de vida nativo |

El frontend no accede a SQL. Los tipos externos o específicos del driver no definen el dominio. Los bindings se generan desde Go, se versionan y se verifican en CI junto con sus consumidores TypeScript.

## Modelo local

Una obra tiene identidad propia y puede estar relacionada con una serie y posición. Una edición conserva formato, idioma, identificador opcional y propiedad. El consumo se registra en ciclos separados, con estado y fechas conocidas. Favorito y notas pertenecen al registro personal.

Libros y Novelas ligeras comparten el modelo de lectura, con un discriminador explícito `books` / `light-novels`. Los servicios se acotan por petición, sin un «medio actual» mutable en el backend. El adaptador exige medio en consultas y mutaciones; un ID de otra sección no habilita leerlo, editarlo o borrarlo desde el contexto equivocado. La edición no cambia el medio de una entrada existente. Las series también tienen identidad por medio, aunque coincidan sus nombres.

La interfaz inicial trabaja con una edición/ciclo por entrada de lectura. El resguardo preserva todas las entidades almacenadas, no solo la proyección elegida por un listado. Las fechas desconocidas no se reconstruyen a partir de la fecha de importación o respaldo.

## Persistencia y recuperación

SQLite se encapsula detrás de operaciones concretas. Sus transacciones evitan actualizaciones parciales. El esquema se versiona y las migraciones existentes generan una copia previa cuando corresponda. Un esquema futuro no reconocido se rechaza.

El resguardo produce SQLite y un modelo JSON independiente. Ambas representaciones provienen del mismo snapshot consistente. Las generaciones se preparan temporalmente y se publican completas; las fallidas no sustituyen la última válida.

Restaurar valida el archivo, prepara una base nueva, conserva una copia de la biblioteca activa y serializa el reemplazo frente a operaciones de biblioteca. Los archivos de configuración de la máquina no forman parte de los datos recuperados. Cambiar de motor requeriría otro adaptador y migración, pero el JSON no obliga a leer primero SQLite.

## Extensión por medios y proveedores

Las secciones tienen identidad y reglas propias. Compartir mecanismos no implica imponer la misma estructura a todos los medios. Importar archivos externos, vincular metadata y restaurar respaldos son operaciones distintas.

Libros usa Hardcover como catálogo principal. Open Library complementa campos faltantes de la edición seleccionada mediante ISBN seguro; no se promueve automáticamente a principal ni se consulta por cada candidato. El snapshot conserva referencias, fuentes y advertencias por separado del registro personal. Refrescar no pisa título, autoría, serie, notas, estado ni fechas; un fallo conserva el snapshot anterior.

La selección se prepara en Go y se confirma mediante un token temporal, sin aceptar un snapshot inventado desde el frontend. Las coincidencias locales exigen revisión antes de vincular; una edición ya vinculada se abre sin duplicarla. El catálogo remoto no está habilitado para Novelas ligeras. Los importadores externos siguen pendientes.

`BookSearch` presenta primero coincidencias locales y después candidatos remotos en el mismo flotante. `BookProfile` comparte la presentación de fichas locales y remotas; la gestión depende de si existe registro personal. Go reconoce coincidencias antes del alta y prioriza la edición local o la más informativa al abrir una obra. Cambiar de edición se gestiona desde la propia ficha. La configuración de claves queda en el encabezado, separada del contenido bibliográfico.

## Distribución

El frontend está embebido en el ejecutable. La identidad de build incluye versión, revisión y perfil; puede consultarse sin abrir SQLite. Las publicaciones contienen el ejecutable validado, no una biblioteca ni configuración personal. Ver [desarrollo](development.md), [formato](backup-format.md) y [releases](releases/README.md).
