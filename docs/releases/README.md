# Publicaciones y artefactos

## Eventos

| Evento | Resultado |
| --- | --- |
| Push a `dev` | Sin automatización remota |
| PR dirigido a `main` | Validación Windows y artefacto portable con perfil `test` |
| Push a `main` después de integrar | Validación, artefacto `habitual` y publicación automática si pasa |

El merge a `main` constituye aprobación de publicación. Los checks obligatorios de la rama impiden integrar si falla el contrato requerido. El primer artefacto de PR debe probarse antes de aprobar la primera publicación.

Actualizar `dev` cuando participa en un PR abierto a `main` vuelve a validar ese PR mediante su evento `pull_request`. No existe un workflow de push a `dev` ni se publica por esa actualización.

## Descargar una prueba de PR

Un artefacto es el resultado de una ejecución de GitHub Actions, no una instalación ni una release pública:

1. Abrir el PR y su check **Windows x64**.
2. Entrar al resumen de la ejecución de Actions.
3. En **Artifacts**, descargar **Trove-windows-x64**.
4. Extraer el ZIP completo y abrir `Trove.exe`.

Ese artefacto usa `data/test/`. Incluye `build-info.json`, versión/revisión consultables, instrucciones y checksums. Es necesario WebView2 Runtime para abrir la ventana; no se necesita Go, Node o SQLite instalado. La descarga de artefactos puede requerir iniciar sesión en GitHub.

## Versión y contenido

Las integraciones en `main` reciben `0.1.<número de ejecución>` y tag `v0.1.<número>`. Los PR reciben una identidad `0.0.0-pr.<PR>.<ejecución>`, sin publicación. Los números pueden tener saltos; no representan cambios del esquema de datos.

La versión y SHA se incrustan antes de probar el ejecutable. El job de publicación descarga ese artefacto, verifica identidad/checksums y empaqueta sus archivos; no recompila la aplicación. La release contiene el ZIP Windows x64 y `SHA256SUMS.txt`.

El flujo crea una release borrador, sube los archivos y luego la publica. Reintentar el mismo run reutiliza su versión: una release publicada del mismo commit no se sobrescribe; una colisión de tag con otro commit falla. No se necesitan PAT personales ni commits automáticos de versionado: se usa `GITHUB_TOKEN` con permisos limitados al job.

## Compatibilidad

Versión de app, esquema SQLite y formato JSON son independientes. Una versión anterior puede no abrir una base migrada por otra posterior; recuperar exige datos compatibles. Las notas de publicación deben describir cambios de compatibilidad cuando existan.

Los paquetes se generan desde archivos explícitos. No incluyen bases, backups, preferencias personales, credenciales ni documentación interna. La configuración de Actions o reglas se puede mantener en PRs posteriores; las ejecuciones normales no requieren nuevas autorizaciones manuales de publicación.
