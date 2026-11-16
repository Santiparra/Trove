# Backend, persistencia y recuperación

SQLite embebido, resguardo automático y restauración están implementados. «Backend» significa lógica local de la aplicación, no un servidor separado. El formato abierto está documentado en [formato de respaldo](backup-format.md).

## Persistencia activa

Trove abre y gestiona su base local, migraciones y conexiones. SQLite no requiere instalar un servidor. Dependiendo del modo puede utilizar archivos auxiliares; copiar únicamente el archivo principal mientras se escribe no garantiza un respaldo completo.

Los datos activos y la carpeta sincronizada de respaldo son ubicaciones distintas. Desarrollo, prueba y uso habitual tienen directorios independientes. La copia local previa a una migración ya implementada no sustituye el sistema cotidiano de resguardo.

## Dos representaciones de respaldo

Cada generación terminada incluye:

| Representación | Función |
| --- | --- |
| SQLite consistente | Recuperación nativa de Trove sobre SQLite, conservando el estado incluido en el resguardo |
| JSON estructurado y versionado | Recuperación lógica independiente de SQLite, conservando identidades, relaciones y significado de los datos |

Son archivos de salida, no los archivos sobre los que trabaja la app. La carpeta se configura una vez por perfil y puede ser sincronizada por Drive/Dropbox; no se integran sus APIs.

El JSON describe conceptos de Trove —obras, ediciones, series, ciclos y referencias—, no sentencias SQL, nombres de tablas ni tipos específicos del driver. Su versión de formato es independiente de la versión de aplicación y del esquema SQLite. CSV no es la representación principal del respaldo completo por sus limitaciones para relaciones y estructuras anidadas.

Cambiar de motor en el futuro requerirá implementar el adaptador/importador que reconstruya ese modelo sobre el nuevo almacenamiento. El JSON evita necesitar SQLite como intermediario, pero no es una conversión automática a cualquier motor ni un estándar que otras aplicaciones deban entender. El respaldo nativo sí requiere soporte de SQLite para interpretarlo.

## Contenido

- Identidades locales y referencias externas con origen/tipo.
- Títulos, autoría y datos propios de edición disponibles.
- Series, relaciones y posiciones.
- Estado, notas, fechas conocidas y ciclos de consumo.
- Favorito y propiedad, sin perder su asociación con obra/edición.
- Otros datos livianos/textuales incluidos en el alcance implementado. La metadata textual que Trove obtenga posteriormente podrá conservarse según sus reglas; no se importa ahora metadata secundaria de archivos externos.

Excluir portadas y recursos pesados, cachés del WebView, logs, credenciales y configuración de rutas de la máquina. Conservar referencias útiles para recuperar recursos ausentes. Restaurar no obliga a consultar proveedores ni a que esos recursos sigan disponibles.

## Generación automática acordada

1. Cada cambio confirmado deja el resguardo pendiente. Antes de configurar carpeta, se informa que falta configuración; los cambios locales no se bloquean.
2. Tras dos segundos sin nuevos cambios, preparar una generación; agrupar escrituras y no publicar otra si el contenido coincide con la última generación completa. Al iniciar, comprobar el contenido para recuperar pendientes de un cierre forzado.
3. Al cierre normal, completar lo pendiente. Una interrupción forzada puede impedir ese último resguardo; la persistencia activa no depende exclusivamente de él.
4. Crear ambas representaciones a partir del mismo estado consistente, con identificador de generación compartido. No exportar JSON desde un estado distinto al del SQLite correspondiente.
5. Preparar y validar archivos temporales antes de publicar los nuevos archivos. Conservar última generación válida y una anterior.
6. Mostrar estado al día, pendiente o error. Un fallo de respaldo no debe presentarse como pérdida o fallo del cambio ya guardado localmente.

La salida está separada por perfil (`development`, `test`, `habitual`) y luego por generación. Se escribe primero en un directorio temporal y se publica renombrándolo cuando sus archivos están terminados: `trove.sqlite`, `trove.json` y un pequeño `complete.json` con identidad y checksums. Se conservan dos generaciones completas. Carpetas ajenas, incompletas o con checksum inválido no se usan como última copia ni se borran indiscriminadamente.

El cliente de nube puede sincronizar los archivos por separado. Cada representación validada puede restaurarse por sí sola; el manifiesto sirve para publicación/retención local, no es requisito para recuperar un archivo aislado. El estado «al día» confirma archivos locales, no subida a la nube. Los fallos mantienen el pendiente y se reintentan; la carpeta vuelve a validarse antes de escribir. No se añade un servicio de sincronización propio.

No hay cifrado propio ni login de Trove para este alcance; ver [seguridad](security.md).

## Restauración acordada

Una sola acción **Restaurar respaldo** acepta SQLite o JSON de Trove. Para esta entrega reemplaza la biblioteca completa del perfil seleccionado; no mezcla bibliotecas ni restaura medios individuales.

1. Elegir archivo y validar identidad del formato, versión compatible, integridad y relaciones.
2. Mostrar información del respaldo y destino; pedir confirmación del reemplazo.
3. Construir la base recuperada en una ubicación temporal. El SQLite identificado de Trove se lee en modo solo lectura, comprobando integridad, referencias y esquema compatible (1 o 2). Ambos formatos pasan al modelo lógico validado y reconstruyen el esquema actual conservando IDs y referencias, sin ejecutar triggers o SQL aportado por el archivo. Una base SQLite arbitraria o la base activa sin identificación de respaldo no se aceptan como copia propia.
4. Validar la base preparada antes de tocar la activa y crear una copia de seguridad de la biblioteca actual.
5. Cerrar conexiones y reemplazar la base activa, con recuperación de la anterior ante fallos. No eliminar la única copia válida para intentar instalar otra. La copia previa `before-restore-*.sqlite` queda en el directorio del perfil y es restaurable por sí sola. La confirmación usa un token de preparación que vence a los diez minutos y un checksum para impedir instalar una copia temporal modificada.
6. Reabrir y actualizar la interfaz. La biblioteca queda utilizable offline; metadata y portadas ausentes se recuperarán mediante otra operación futura.

Restaurar no crea nuevos ciclos, no asigna fechas de hoy a consumo antiguo ni ejecuta automatismos de transición de estado. No importar del respaldo rutas de salida o credenciales que hagan que el perfil de prueba escriba sobre el habitual.

La restauración se serializa frente a las operaciones de biblioteca, para evitar un guardado que lea la base anterior y escriba sobre la nueva. El cierre normal espera los guardados ya aceptados, rechaza nuevas escrituras y completa el respaldo pendiente antes de cerrar SQLite. Las copias de seguridad de reemplazo se conservan localmente; no forman parte de la retención de dos generaciones sincronizables.

## Diferencia respecto de importaciones externas

Importar archivos de StoryGraph, MAL, Trakt o IMDb incorpora información con revisión, clasificación, emparejamiento y prioridad local/entrante. Restaurar recupera un estado de Trove y no usa ese gestor de conflictos. Ambas representaciones propias comparten el flujo de restauración, no son dos funcionalidades de importación de catálogos.

La reconstrucción futura de metadata por medio/proveedor sigue siendo independiente de restaurar datos propios por secciones.

## Comprobaciones de esta entrega

Probar ida y vuelta por ambas representaciones en un perfil vacío, equivalencia del contenido, preservación de relaciones y fechas desconocidas, archivos inválidos/incompatibles, generaciones incompletas, fallo al generar y recuperación ante fallo de reemplazo. Usar datos sintéticos y SQLite real, sin depender de proveedores ni de Drive/Dropbox para las pruebas automáticas.
