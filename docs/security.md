# Seguridad y privacidad

Estado: privacidad entendida como alcance local y mínima dependencia de cuentas y servicios. Sin login ni cifrado propio de Trove para la base o los respaldos en el alcance inicial. Implementación de controles pendiente.

## Datos locales y tráfico externo

La biblioteca vive localmente, pero una llamada a un proveedor sí transmite información: consulta de búsqueda, ID de obra, dirección IP visible para el servicio y, si corresponde, credenciales o tokens. HTTPS protege el transporte, no oculta la consulta al proveedor.

Política acordada: sin telemetría, analítica, envío de biblioteca ni sincronización de estados/notas con servicios externos. Los adaptadores enviarán solo lo necesario para la operación solicitada. Buscar metadata no debe enviar notas, puntajes o historial personal. Un futuro envío de esos datos requeriría una decisión explícita de producto.

Preferir consultas sin cuenta cuando el proveedor lo permita. Evaluar autenticación caso por caso, no asumir que toda API exige login. No solicitar contraseñas de IMDb, MyAnimeList u otros servicios para operaciones que no las necesitan; una API key de aplicación y la autenticación de una cuenta personal son requisitos distintos. El usuario obtiene las exportaciones con herramientas externas y Trove importa sus archivos; no necesita autenticarse para leer esos archivos.

El cliente de Drive/Dropbox transmite los respaldos por su cuenta; es distinto del tráfico iniciado por Trove. Incluir datos personales en esos respaldos implica confiarlos al almacenamiento configurado.

## Red: propuesta de destinos permitidos

- Denegar por defecto las conexiones de red de la aplicación y permitir HTTPS solo a destinos declarados de los proveedores configurados, incluidas APIs, autenticación y CDN de imágenes necesarios.
- Centralizar solicitudes en adaptadores controlados desde el núcleo; evitar que la interfaz cargue scripts, páginas, fuentes o imágenes remotas arbitrarias.
- Comprobar esquema, host y puerto con análisis de URL, no por coincidencia de texto; validar cada redirección y bloquear destinos locales/privados no autorizados. No aceptar automáticamente cualquier URL recibida del proveedor.
- Configuración de destinos explícita y limitada: no ampliar permisos silenciosamente porque una respuesta redirige a otro dominio.
- Autenticación en navegador externo, si se requiere, con flujo específico por definir. Abrir enlaces elegidos por el usuario no equivale a permitir navegación remota dentro de la app.

Esta política controla las rutas de red implementadas por Trove, no reemplaza un firewall ni garantiza controlar el tráfico del sistema, del navegador externo o de un runtime que se actualiza por separado. Revisar las capacidades reales del contenedor al elegir stack. No diseñar un motor de plugins ni un panel de permisos complejo sin necesidad.

## Credenciales y portabilidad

Propuesta: tokens fuera de logs, HTML compartidos y respaldos ordinarios; evitar credenciales en URLs. Utilizar almacenamiento protegido por Windows cuando corresponda.

La protección ligada a una cuenta/PC puede impedir trasladar tokens a otra instalación. Prioridad sugerida: biblioteca portable y volver a autenticar proveedores al cambiar de PC. Un secreto distribuido dentro de un ejecutable de escritorio no puede tratarse como un secreto confidencial de servidor.

## Datos externos y archivos

- Validar estructura y tamaños de importaciones, respuestas y recursos descargados; establecer tiempos de espera y límites.
- Tratar texto externo como datos: no ejecutar HTML o scripts de sinopsis, nombres o notas. Escapar contenido al generar el HTML de recomendaciones.
- Si se implementan paquetes de respaldo, impedir que rutas de sus entradas escriban fuera de la carpeta de restauración.
- Limitar la comunicación UI/núcleo a operaciones concretas; no exponer ejecución arbitraria de comandos ni acceso irrestricto al sistema de archivos desde contenido remoto.
- Preservar la última copia válida ante fallos y validar restauraciones antes de sustituir datos activos.

## Pendientes de producto

Concretar política de destinos y actualización, contenido predeterminado del HTML compartido y presentación de errores de autenticación/red. El cifrado propio no es un pendiente bloqueante: se decidió no incluirlo inicialmente.

## Protección local y respaldos: decisión confirmada

El usuario considera suficiente la protección del entorno Windows y de su almacenamiento para una biblioteca con puntajes y opiniones. No se implementan cuentas, contraseñas de acceso a Trove, cifrado de la base ni cifrado de backups gestionado por la app en el alcance inicial.

Los archivos de datos y resguardo serán legibles para quien obtenga acceso a ellos; no se presume que Windows o el proveedor de nube los cifren con una clave exclusiva del usuario. Este límite se acepta sin añadir gestión de claves ni pasos de desbloqueo a la recuperación cotidiana.

Esto no elimina el cuidado de credenciales de proveedores ni las reglas sobre tráfico y contenido externo. Un eventual cifrado adicional requeriría una nueva decisión del usuario; la recomendación anterior de cifrar por defecto las copias sincronizadas queda sustituida por este acuerdo.
