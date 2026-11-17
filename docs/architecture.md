# Arquitectura

Trove es una aplicación modular local: interfaz React/TypeScript en Wails v3, casos de uso Go y SQLite embebido. No requiere servidor de base de datos ni cuenta propia.

## Responsabilidades

| Componente | Responsabilidad |
| --- | --- |
| `frontend/src/books` | Presentación, formularios, navegación y estado de interacción |
| `frontend/src/backup` | Configuración visible, estado y confirmación de recuperación |
| `internal/desktop` | Adaptadores Wails, contratos de entrada/salida y diálogos nativos |
| `internal/books` | Identidad, validaciones y casos de uso de biblioteca |
| `internal/storage/sqlite` | SQL, transacciones, migraciones y adaptación de snapshots |
| `internal/backup` | Formato lógico, validación, generaciones y restauración mediante un contrato de backend |
| `internal/config` | Resolución de perfiles y ubicación de datos |
| `main.go` | Composición de dependencias y ciclo de vida nativo |

El frontend no accede a SQL. Los tipos externos o específicos del driver no definen el dominio. Los bindings se generan desde Go, se versionan y se verifican en CI junto con sus consumidores TypeScript.

## Modelo local

Una obra tiene identidad propia y puede estar relacionada con una serie y posición. Una edición conserva formato, idioma, identificador opcional y propiedad. El consumo se registra en ciclos separados, con estado y fechas conocidas. Favorito y notas pertenecen al registro personal.

La interfaz inicial trabaja con una edición/ciclo por libro. El resguardo preserva todas las entidades almacenadas, no solo la proyección elegida por un listado. Las fechas desconocidas no se reconstruyen a partir de la fecha de importación o respaldo.

## Persistencia y recuperación

SQLite se encapsula detrás de operaciones concretas. Sus transacciones evitan actualizaciones parciales. El esquema se versiona y las migraciones existentes generan una copia previa cuando corresponda. Un esquema futuro no reconocido se rechaza.

El resguardo produce SQLite y un modelo JSON independiente. Ambas representaciones provienen del mismo snapshot consistente. Las generaciones se preparan temporalmente y se publican completas; las fallidas no sustituyen la última válida.

Restaurar valida el archivo, prepara una base nueva, conserva una copia de la biblioteca activa y serializa el reemplazo frente a operaciones de biblioteca. Los archivos de configuración de la máquina no forman parte de los datos recuperados. Cambiar de motor requeriría otro adaptador y migración, pero el JSON no obliga a leer primero SQLite.

## Extensión por medios y proveedores

Las secciones tienen identidad y reglas propias. Compartir mecanismos no implica imponer la misma estructura a todos los medios. Importar archivos externos, vincular metadata y restaurar respaldos son operaciones distintas.

Las integraciones de proveedores e importadores externos aún no están implementadas. Su incorporación debe mantener independientes los registros propios y extender explícitamente formatos, contratos y compatibilidad. No se implementa un framework genérico de plugins.

## Distribución

El frontend está embebido en el ejecutable. La identidad de build incluye versión, revisión y perfil; puede consultarse sin abrir SQLite. Las publicaciones contienen el ejecutable validado, no una biblioteca ni configuración personal. Ver [desarrollo](development.md), [formato](backup-format.md) y [releases](releases/README.md).
