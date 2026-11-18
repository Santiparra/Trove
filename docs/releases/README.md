# Publicaciones y artefactos

## Historial de novedades

| Publicación | Novedades |
| --- | --- |
| Corte actual, posterior a v0.1.4; número asignado por Actions | [Catálogo de Libros y biblioteca de Novelas ligeras](catalogo-libros-y-novelas.md) |
| [v0.1.4](https://github.com/Santiparra/Trove/releases/tag/v0.1.4) | Entrega anterior; conserva sus notas originales en GitHub |

Los documentos de cada corte están orientados al uso de la aplicación, no a enumerar commits. Permanecen en esta carpeta como historial. La versión exacta y fecha de publicación se consultan en [GitHub Releases](https://github.com/Santiparra/Trove/releases).

## Preparar las notas de cada corte

1. Crear un documento nuevo, por ejemplo `docs/releases/catalogo-libros-y-novelas.md`, con novedades, mejoras y límites relevantes para usuarios. Conservar los documentos de cortes anteriores.
2. Escribir su nombre en `docs/releases/current.txt`, sin rutas ni prefijos. Ese archivo selecciona el documento que se publicará; no contiene un número de versión.
3. Agregar el documento al historial de esta página y actualizar el resumen del README y las guías afectadas. Al preparar el siguiente corte, identificar la versión ya publicada en la fila anterior y agregar una fila nueva.
4. Incluir los archivos en el PR a `main`. La verificación rechaza una selección inexistente, vacía o sin título. Antes del merge, revisar el check Windows y el artefacto correspondiente al último commit del PR.

El job de publicación usa el documento seleccionado como cuerpo de la release mediante `--notes-file`. También lo aplica al reanudar un borrador. No sustituye las notas de una release que ya fue publicada correctamente. No usa las notas autogeneradas de GitHub ni depende de copiar el cuerpo del PR o del mensaje de merge.

El título y cuerpo del PR sirven para revisar el conjunto de cambios; el título y cuerpo de la confirmación de merge pasan a la historia de Git. Pueden resumir las mismas novedades, pero **la fuente de las notas públicas de la release es el documento seleccionado**. El título de la release sigue siendo `Trove v<versión>`.

## Eventos

| Evento | Resultado |
| --- | --- |
| Push a `dev` | Sin automatización remota |
| PR dirigido a `main` | Validación Windows y artefacto portable con perfil `test` |
| Push a `main` después de integrar | Validación, artefacto `habitual` y publicación automática si pasa |

El merge a `main` constituye aprobación de publicación. Los checks obligatorios de la rama impiden integrar si falla el contrato requerido. El artefacto del último commit del PR permite comprobar el mismo contrato de distribución que se validará al publicar.

La integración usa merge commits para conservar la relación entre las ramas persistentes; no elimina `dev` después del merge. Antes de iniciar el siguiente bloque de trabajo conviene actualizar `dev` desde `main` para mantener los checks sobre una base actualizada.

Actualizar `dev` cuando participa en un PR abierto a `main` vuelve a validar ese PR mediante su evento `pull_request`. No existe un workflow de push a `dev` ni se publica por esa actualización.

## Descargar una prueba de PR

Un artefacto es el resultado de una ejecución de GitHub Actions, no una instalación ni una release pública:

1. Abrir el PR y su check **Windows x64**.
2. Entrar al resumen de la ejecución de Actions.
3. En **Artifacts**, descargar **Trove-windows-x64**.
4. Extraer el ZIP completo y abrir `Trove.exe`.

Ese artefacto usa `data/test/`. Incluye `build-info.json`, versión/revisión consultables, instrucciones y checksums. Es necesario WebView2 Runtime para abrir la ventana; no se necesita Go, Node o SQLite instalado. La descarga de artefactos puede requerir iniciar sesión en GitHub.

## Versión y contenido

Las integraciones en `main` reciben `0.1.<número de ejecución>` y tag `v0.1.<número>`. Los PR reciben una identidad `0.0.0-pr.<PR>.<ejecución>`, sin publicación. Los números pueden tener saltos; no representan cambios del esquema de datos. Un PR y el push posterior a `main` son ejecuciones distintas: que la última release sea `0.1.4` no implica que la siguiente sea `0.1.5`. No se fija por adelantado el número en el documento de novedades.

La versión y SHA se incrustan antes de probar el ejecutable. El job de publicación descarga ese artefacto, verifica identidad/checksums y empaqueta sus archivos; no recompila la aplicación. La release contiene el ZIP Windows x64 y `SHA256SUMS.txt`.

El flujo crea una release borrador, sube los archivos y luego la publica. Reintentar el mismo run reutiliza su versión: una release publicada del mismo commit no se sobrescribe; una colisión de tag con otro commit falla. No se necesitan PAT personales ni commits automáticos de versionado: se usa `GITHUB_TOKEN` con permisos limitados al job.

## Compatibilidad

Versión de app, esquema SQLite y formato JSON son independientes. Una versión anterior puede no abrir una base migrada por otra posterior; recuperar exige datos compatibles. Las notas de publicación deben describir cambios de compatibilidad cuando existan.

Los paquetes se generan desde archivos explícitos. No incluyen bases, backups, preferencias personales, credenciales ni documentación interna. La configuración de Actions o reglas se puede mantener en PRs posteriores; las ejecuciones normales no requieren nuevas autorizaciones manuales de publicación.
