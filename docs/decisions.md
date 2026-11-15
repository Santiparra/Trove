# Decisiones y preguntas abiertas

Fuente: conversación de planificación con el usuario, 2026-11-15. Este registro no convierte las propuestas del asistente en decisiones aprobadas.

## Confirmado por el usuario

| ID | Decisión | Consecuencia |
| --- | --- | --- |
| D01 | Trove es el nombre provisional | Puede cambiar sin condicionar el diseño |
| D02 | Una persona, una PC; sin acceso móvil/remoto | Priorizar experiencia local y arranque autónomo |
| D03 | Datos propios editables y funcionamiento independiente de proveedores | La metadata externa no puede ser requisito para gestionar la biblioteca |
| D04 | Importadores y proveedores de metadata pueden ser distintos | No acoplar el rescate de un listado con el proveedor de búsqueda |
| D05 | Temporadas de anime como entradas separadas vinculadas | Conservar relaciones de secuela/precuela y agrupación |
| D06 | Película y remake son obras diferentes | No fusionarlas por título o relación |
| D07 | Libros y volúmenes de novelas ligeras como entradas separadas vinculadas a su serie | Registrar consumo individual y progreso del conjunto |
| D08 | Un libro puede tener varias ediciones; la lectura corresponde a una edición | Separar identidad de obra y edición |
| D09 | Adaptaciones entre medios son registros independientes relacionados | Un anime no comparte registro de consumo con su manga o novela |
| D10 | Relectura como estado visible propio | Falta acordar historial y repetición del consumo |
| D11 | Registro de ítems individuales y seguimiento del progreso de la serie | Definir cómo se calcula y si se permite modificar el agregado |
| D12 | Exportar texto para usar en IA externa, sin integrar IA | No se requieren modelos, claves ni cuentas de IA |
| D13 | Exportar selecciones como HTML estático para amigos | Definir contenido, privacidad y empaquetado |
| D14 | Por ahora, solo planificación y documentación | No implementar ni preparar entorno |
| D15 | Primer medio: libros; primera fuente de importación: The StoryGraph | Biblioteca de más de 250 ítems según el usuario; formato y campos reales por comprobar |
| D16 | SQLite aceptado con resguardo automático y exportación abierta de fallback | No se necesita edición externa en LibreOffice; biblioteca embebida, sin servicio administrado |
| D17 | Respaldos ligeros, sin metadata externa ni portadas; conservar referencias para recuperarlas | Restaurar datos propios independientemente de proveedores; la recuperación externa es de mejor esfuerzo |
| D18 | Resguardo sin pasos manuales cotidianos y carpeta de salida configurable | Trove prepara archivos consistentes y el cliente de Drive/Dropbox los sincroniza; política exacta pendiente |
| D19 | Experiencia portable preferida | Evaluar dependencias del contenedor y permitir configurar las salidas de respaldo |
| D20 | Cambios siempre locales; el usuario hace commits y publicaciones | No hacer commits, push ni publicar; no alterar su staging |
| D21 | Consultar decisiones importantes de producto no cubiertas | No cerrar unilateralmente comportamientos nuevos |

## Propuestas pendientes de aprobación

- Identidad interna propia, separada de IDs externos.
- Separar obra, relación personal con la obra y metadata externa.
- Núcleo modular y adaptadores para persistencia, importaciones y proveedores.
- Conservar el último snapshot válido y no pisar cambios personales al actualizar.
- Importaciones con vista previa, repetibles sin duplicados y sin sobrescritura silenciosa.
- Registrar desde temprano el historial mínimo necesario para fechas, relecturas y estadísticas.
- Interfaz React + TypeScript + Vite; contenedor y backend todavía por definir.

## Preguntas prioritarias

1. ¿Un único par de archivos de resguardo para todos los medios o un par por medio? El usuario sugirió lo segundo como ejemplo; no está aprobado un particionado físico de la base.
2. ¿Qué mínimo identificativo conservar además de las referencias externas para reconocer obras y ediciones sin proveedor? Propuesta: título, autor e identificadores de edición disponibles.
3. ¿Qué frecuencia y comportamiento ante fallos tendrá el resguardo automático? Propuesta: tras cambios con una breve espera y al cierre normal; no depender solo del cierre.
4. ¿Go con Wails, Electron con TypeScript o Tauri con Rust? Ver [comparación](tech-stack.md).
5. ¿Qué representa una serie y cómo se calcula su progreso si faltan volúmenes, sigue publicándose o se consume fuera de orden?
6. ¿Cómo distinguir una relectura en curso de una terminada, conservando el consumo anterior? ¿Misma regla para revisualizaciones?
7. ¿Qué datos mínimos, estados, escala de puntaje y unidad de progreso necesita el primer medio?
8. ¿Qué diferencias entre ediciones importan: idioma, traducción, ISBN, editorial, formato, audiolibro?

## Preguntas posteriores

- ¿Qué campos manuales prevalecen sobre metadata? ¿Se distingue título propio de título del proveedor?
- ¿Cómo resolver duplicados y discrepancias entre distintas importaciones?
- ¿Qué fechas existen, con qué precisión, y qué ocurre con fechas desconocidas?
- ¿Qué contenido incluir por defecto en HTML compartido? ¿Portadas embebidas y funcionamiento sin Internet?
- ¿Cuánto historial de backups conservar sin multiplicar innecesariamente archivos? El contenido de resguardo se delimita en [backend](backend.md).
- ¿Cómo registrar manga/manhwa y progreso por capítulos o volúmenes?
- ¿Qué proveedores ofrecen la cobertura necesaria y qué limitaciones reales tienen sus APIs y exportaciones?
