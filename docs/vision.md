# Visión y alcance

## Problema

El seguimiento de consumo está distribuido entre varios servicios. Trove busca centralizar y conservar la biblioteca, el historial, los puntajes y las notas personales aunque desaparezcan esos servicios.

## Requisitos confirmados

- Uso personal en una única PC, sin acceso remoto ni desde el teléfono.
- Abrir la aplicación debe ser suficiente: evitar administración manual de servicios o de una base de datos.
- No requerir una cuenta de Trove. Algunas integraciones podrían necesitar credenciales del proveedor; los proveedores aún no están elegidos.
- Desarrollo progresivo, feature a feature, con prácticas profesionales proporcionales al proyecto.
- Gestionar libros, anime, manga, novelas ligeras, manhwa, series de TV y películas, con secciones especializadas.
- Crear y editar datos personales sin depender de proveedores.
- Importar listados e información personal de servicios actuales, conservando los datos utilizables aunque no se consiga metadata adicional.
- Vincular obras con proveedores, con posibilidad de aprobación y corrección manual.
- Conservar metadata local y actualizarla explícitamente mediante snapshots.
- Facilitar el resguardo y la recuperación después de formatear la PC, con SQLite y exportación abierta de fallback, sin edición externa obligatoria.
- Resguardo automático ligero en una carpeta configurable sincronizable con Drive/Dropbox, sin tareas manuales cotidianas. Incluir datos propios y contenido liviano/textual; excluir portadas y otros recursos pesados, conservando referencias para recuperarlos.
- Separar datos locales propios de metadata externa. Cada medio tiene particularidades y debe permitir reconstruir metadata por sección o proveedor sin afectar a los demás.
- Experiencia portable preferida; detalles de distribución por evaluar.
- Exportar un listado de consumo reciente para copiarlo en una IA externa. No se solicita integración de IA dentro de Trove.
- Generar un HTML estático de obras seleccionadas, con metadata, para compartir recomendaciones con amigos. Formato y portabilidad pendientes.

## Fuentes actuales

| Servicio | Uso actual |
| --- | --- |
| MyAnimeList | Anime, manga y novelas ligeras populares |
| The StoryGraph | Libros y algunas novelas ligeras |
| Trakt | Series de TV y películas |
| IMDb | Puntajes de series y películas, sin listados personales |

No existe una fuente actual dedicada a manhwa. Las películas tienen menor prioridad. El primer medio será libros, con The StoryGraph como primera fuente de importación. El gestor permitirá separar manualmente los medios mezclados en un archivo.

## Funcionalidades previstas para etapas posteriores

Filtros y ordenación, progreso, fechas de consumo, historial, estadísticas generales y por género. Su alcance y orden exacto siguen abiertos.

## Fuera del alcance actual

Implementación de la aplicación, acceso móvil/remoto, cuentas de Trove e integración de modelos de IA.
