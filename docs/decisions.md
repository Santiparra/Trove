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
| D15 | Primer medio: libros; primera fuente de importación: The StoryGraph | Formato examinado; mapeo preciso a concretar al implementar |
| D16 | SQLite aceptado con resguardo automático y exportación abierta de fallback | No se necesita edición externa en LibreOffice; biblioteca embebida, sin servicio administrado |
| D17 | Respaldos con datos propios y contenido liviano/textual; excluir portadas y otros recursos pesados | Aclaración posterior: no excluir toda la metadata por ser externa; conservar referencias para recuperar recursos ausentes |
| D18 | Resguardo sin pasos manuales cotidianos y carpeta de salida configurable | Trove prepara archivos consistentes y el cliente de Drive/Dropbox los sincroniza; política exacta pendiente |
| D19 | Experiencia portable preferida | Evaluar dependencias del contenedor y permitir configurar las salidas de respaldo |
| D20 | Cambios siempre locales; el usuario hace commits y publicaciones | No hacer commits, push ni publicar; no alterar su staging |
| D21 | Consultar decisiones importantes de producto no cubiertas | No cerrar unilateralmente comportamientos nuevos |
| D22 | Dos capas de datos: registro local independiente y metadata vinculada a proveedores | Título local, referencias externas, estado, notas, fechas y progreso pertenecen al registro propio; actualizar metadata no debe sustituirlo |
| D23 | Medios moderadamente independientes, con estructura y metadata especializadas | Reconstruir metadata por medio o proveedor sin modificar datos de otros medios |
| D24 | El número de archivos no es una prioridad; almacenamiento interno y exportación son decisiones distintas | Tablas diferentes no exigen archivos SQLite diferentes; diseño físico pendiente |
| D25 | Cada medio opera desde su pestaña con tipo interno, estructura propia y proveedor de metadata vinculado, por definir | El medio determina el adaptador y los registros objetivo; no consultar otros medios ni resolver globalmente el alcance por cobertura del proveedor |
| D26 | Buscar y consultar detalles no agrega registros a la biblioteca ni al respaldo | Solo agregar explícitamente a un listado promueve la información a datos persistidos de biblioteca; caché temporal posible |
| D27 | Importación con adaptador específico por origen, separando datos propios de metadata secundaria | Emparejar con registros locales y ofrecer políticas de actualización de duplicados; semántica exacta pendiente |
| D28 | Refresco/reconstrucción de metadata como operación repetible sobre el listado interno existente | Botón explícito por sección, independiente de importar y de buscar; conservar identidad y datos propios |
| D29 | Al importar, elegir conservar existentes o sobrescribirlos con el origen | Prioridad local: no tocar el registro existente. Prioridad importada: aplicar datos entrantes en campos correspondientes, incluso sobre cambios locales; no requiere historial comparativo de importaciones |
| D30 | El usuario dispara explícitamente la preparación de cada versión | No generar releases automáticamente por feature; proceso y notas en `docs/releases/` |
| D31 | Arquitectura limpia y desacoplada, sin exigir DDD; lógica de negocio en backend y frontend ligero | Aislar también persistencia, reducir duplicación con criterio y seguir convenciones; ver arquitectura |
| D32 | Stack confirmado: Wails v3 + Go + React/TypeScript + Vite + SQLite | Bibliotecas auxiliares y driver concreto se seleccionarán al implementar |
| D33 | Cada lectura y revisualización constituye un ciclo de consumo independiente | Conservar consumos anteriores al iniciar una repetición |
| D34 | Crear obras y ediciones manuales sin proveedor, con vinculación posterior opcional | No encontrar metadata no bloquea ni invalida el registro; vincular no debe perder datos locales |
| D35 | Privacidad prioritaria; evitar autenticación cuando el proveedor permita consultas anónimas | Autenticación solo cuando sea necesaria, caso por caso; sin telemetría ni envío de biblioteca a proveedores |
| D36 | Las exportaciones de servicios las obtiene el usuario con herramientas externas | Trove importa archivos ya creados; no descarga listados de cuentas como parte del alcance acordado |
| D37 | Windows x64, interfaz en español, código en inglés y documentación en español | El idioma de interfaz no traduce títulos; se prevén títulos en inglés y título original complementario |
| D38 | Arquitectura planteada aceptada; primer alcance se define al terminar planificación | Mantener separación de núcleo/adaptadores y no iniciar implementación todavía |
| D39 | Privacidad como alcance local y mínima autenticación; sin login ni cifrado propio de base/backups inicialmente | Se acepta protección del entorno Windows/almacenamiento y archivos legibles para quien acceda a ellos; no añadir gestión de claves |
| D40 | Documentación pública de formatos sin estadísticas ni datos de la biblioteca personal | Originales privados fuera de Git; ejemplos públicos sintéticos |
| D41 | Gestor de importación con revisión, filtros, aprobación, selección de destino en bloque e ignorar entradas | Separar manualmente libros/novelas ligeras y manga/manhwa/novelas ligeras sin depender de metadata online |
| D42 | Recuperar datos propios y referencias de origen, no snapshots de metadata secundaria del archivo importado | Metadata descriptiva se obtendrá del proveedor elegido; descartar perfil, red social, likes y ajustes del servicio |
| D43 | Fechas de origen opcionales; Trove permitirá registrarlas y editarlas para consumo futuro | No inventar fechas ni bloquear importaciones por historial incompleto |
| D44 | Puntajes y precisión definidos por medio considerando su proveedor y el mapeo de importación | No imponer una escala universal ni cambiarla silenciosamente |
| D45 | IMDb posterior a Trakt, como fuente de calificaciones personales | Emparejar y actualizar ratings con prioridad seleccionada, no importar historial inexistente |
| D46 | Workflow aprobado: desarrollo en vivo, build portable de prueba y release solo por indicación del usuario | Perfiles de datos de desarrollo/prueba/habitual separados; convención `0.MINOR.PATCH`; ver releases |
| D47 | Estrategia de testing aprobada: automatización focalizada y validación manual complementaria | Priorizar integridad de datos, SQLite real temporal y fixtures sintéticos; comprobar proporcionalmente cada cambio; ver testing |
| D48 | Primera entrega aprobada: navegación por medios y biblioteca manual de libros con búsqueda local y SQLite | Alcance y aceptación en roadmap; no incluye proveedores, importaciones ni backups automáticos; escala de puntaje pendiente antes de incluirla |
| D49 | Serie y número de orden incluidos en la primera entrega de libros | Compartir identidad de serie entre libros y conservar su posición; no implementar aún tracking agregado avanzado |
| D50 | Implementar en Principal, en este chat | Autorizadas preparación del proyecto, dependencias y comprobaciones de la primera entrega; sin commits ni publicación |
| D51 | Libros abre un portal antes del listado: estados con cantidades, lectura actual, últimos leídos y barra de distribución | Acceso al listado al elegir estado, ver todos o buscar; resumen calculado por backend sobre datos locales |
| D52 | Detalle y formulario de libros en páginas completas, conservando navegación lateral y superior | Sustituir panel lateral y renovar iconos de medios con SVG coherentes |
| D53 | Incorporar eliminación de libro y cambio rápido de estado | Confirmar borrado local de libro y sus datos asociados sin afectar otros libros; cambio de estado no sustituye el resto de campos |
| D54 | Feedback de navegación confirmado: portal compacto, botones de estado, serie sobre título y página local de serie, búsqueda desplegable y notas editables en ficha | Bloque autorizado e implementado; tarjetas nuevas y autores externos siguen fuera del corte |
| D55 | Favorito, propiedad, estado up-next y fechas automáticas de consumo futuro con edición manual | Reglas en dominio; no inventar fechas al importar ni al marcar leído sin inicio; no implementado todavía |
| D56 | Personalización futura: actividad mensual, mejor puntuados del año, favoritos y descubrimiento por autor | Requiere datos y proveedores aún no integrados; no fabricar estadísticas ni aprobar un cálculo de páginas/puntajes implícitamente |
| D57 | «A continuación» es un estado manual; portal adaptable con referencia de escritorio 1920×1080 | No agregar automáticamente libros de una serie; reducir contenido de vista previa y ofrecer acceso al listado completo |
| D58 | Fechas en acciones locales: primer inicio al entrar en leyendo; finalización al entrar en leído solo si hay inicio y falta finalización, sin depender del estado anterior | Conservar fechas existentes, respetar ediciones manuales y validar fin igual o posterior al inicio; no fabricar fechas históricas al importar/restaurar |
| D59 | Resguardo automático aprobado: archivos de salida separados de la base activa, SQLite consistente y JSON lógico independiente del motor | Conservar identidades y relaciones, versión de formato propia y generación compartida; el JSON debe permitir implementar restauración sin SQLite en un motor futuro |
| D60 | Conservar última generación válida y una anterior, con estado visible y carpeta configurable por perfil | Agrupar cambios, completar pendientes al cierre normal y no perder la última copia válida por fallos de generación/sincronización |
| D61 | Una acción de restauración completa de perfil desde SQLite o JSON propios de Trove | Validación, preparación temporal, copia de seguridad previa y confirmación; no mezclar con el gestor de importación de servicios externos |
| D62 | Implementación autorizada de la entrega de resguardo y restauración en Principal, en este chat | Sin importaciones externas, sin cambio de versión de release ni publicación; conservar cambios locales |

## Propuestas pendientes de aprobación

- Formato concreto de la identidad interna propia y sus referencias externas.
- Separar obra, relación personal con la obra y metadata externa.
- Organización concreta del núcleo modular y contratos de adaptadores; el requisito de desacoplamiento ya está confirmado por D31.
- Conservar el último snapshot válido y no pisar cambios personales al actualizar.
- Detalles de detección de coincidencias y mapeo; gestor y prioridad de importación ya confirmados.
- Detalles del historial por ciclo, incluyendo notas y puntajes generales o por consumo; el ciclo independiente ya está confirmado.
- Política de seguridad y red restringida a destinos declarados; ver [seguridad](security.md). La pregunta del usuario se registra como propuesta, no como aprobación de medidas concretas.
- Diseño orientado a continuidad local sin actualizaciones obligatorias por mera existencia de versiones nuevas; ver [durabilidad](longevity.md). No garantiza diez años de compatibilidad ni APIs inmutables.

## Preguntas prioritarias

1. Definir organización física y formato de respaldo según independencia y simplicidad, no según número de archivos. Propuesta: una base activa con núcleo compartido y tablas especializadas; exportación independiente del esquema físico.
2. Además del título local y las referencias externas confirmados, ¿qué campos identificativos propios conservar: autor, ISBN, idioma, etc.? Precisar también qué metadata textual se incluye y límites de tamaño del resguardo.
3. ¿Qué frecuencia y comportamiento ante fallos tendrá el resguardo automático? Propuesta: tras cambios con una breve espera y al cierre normal; no depender solo del cierre.
4. Protección de respaldos resuelta para el alcance inicial mediante D39: sin cifrado propio ni contraseña de Trove. Ver [seguridad](security.md).
5. ¿Qué representa una serie y cómo se calcula su progreso si faltan volúmenes, sigue publicándose o se consume fuera de orden?
6. ¿Cómo distinguir una relectura en curso de una terminada, conservando el consumo anterior? ¿Misma regla para revisualizaciones?
7. ¿Qué datos mínimos, estados, escala de puntaje y unidad de progreso necesita el primer medio?
8. ¿Qué diferencias entre ediciones importan: idioma, traducción, ISBN, editorial, formato, audiolibro?

## Preguntas posteriores

- ¿Qué campos manuales prevalecen sobre metadata? ¿Se distingue título propio de título del proveedor?
- ¿Cómo identificar coincidencias sin confundir obras o ediciones? Precisar mapeos, campos ausentes/vacíos y registros ausentes del archivo. La prioridad entre datos locales e importados ya se elige mediante D29. Ver [integraciones](integrations.md).
- ¿Qué fechas existen, con qué precisión, y qué ocurre con fechas desconocidas?
- ¿Qué contenido incluir por defecto en HTML compartido? ¿Portadas embebidas y funcionamiento sin Internet?
- Detalles técnicos de publicación de generaciones y diagnóstico de resguardo; retención y alcance de restauración ya acordados en D59–D61. Ver [backend](backend.md).
- ¿Cómo registrar manga/manhwa y progreso por capítulos o volúmenes?
- ¿Qué proveedores ofrecen la cobertura necesaria y qué limitaciones reales tienen sus APIs y exportaciones?
