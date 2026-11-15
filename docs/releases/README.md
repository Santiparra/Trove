# Versiones y notas de release

## Regla confirmada

El usuario decide explícitamente cuándo crear una versión. Completar una feature no dispara automáticamente un cambio de versión, compilación de release, nota versionada, tag, commit ni publicación.

Esta carpeta contiene documentación de releases, no ejecutables. No existe todavía una versión de Trove ni un proceso de empaquetado implementado.

## Flujo acordado

El usuario aprobó desarrollo en vivo → build local de prueba → candidato de versión → publicación bajo su control, con datos separados. No se crearon comandos, entornos ni artefactos. La estrategia técnica de comprobación está propuesta en [testing](../testing.md).

### 1. Implementación y revisión en vivo

Trabajar sobre una feature acordada con criterios observables. Comprobar lo modificado con pruebas proporcionales y usar modo desarrollo para revisar interacción en vivo cuando sea necesario o se solicite. Los cambios de frontend podrán aprovechar recarga rápida; los del backend pueden requerir recompilación/reinicio según el tooling. No prometer que todo se actualiza sin reiniciar.

Esto no cambia la versión pública ni genera una nota de release. Los comandos exactos de desarrollo se documentarán al crear la aplicación y verificar su funcionamiento.

### 2. Prueba local empaquetada, sin release

Cuando el usuario pida «preparar una build de prueba», compilar un paquete portable local para ejecutar sin servidor de desarrollo ni editor. Puede agrupar varias features y seguir teniendo como base la versión pública anterior.

Identificarlo claramente como **PRUEBA / NO PUBLICADO**, con identificador de build, revisión de código si existe y marca de cambios locales si corresponde. No depender solo del número de versión pública para distinguirlo. No hacer un commit para obtener ese identificador.

Entregar una lista breve y específica de qué probar y qué se comprobó automáticamente. El usuario ejecuta, prueba y devuelve observaciones; se corrige y se prepara otra build cuando corresponda. Cada modificación invalida la aprobación de la build anterior en las partes afectadas.

### 3. Preparación explícita de una versión

Solo cuando el usuario indique que desea preparar una versión:

1. Acordar alcance y número; si falta, proponerlo sin asignarlo unilateralmente.
2. Actualizar versión y crear `docs/releases/vX.Y.Z.md` con cambios reales y comprobaciones ejecutadas.
3. Generar un candidato local con identidad propia y comprobar el paquete con la versión final prevista, incluyendo migraciones relevantes. No asumir que el paquete definitivo está probado porque funcionó una build previa con otro contenido.
4. El usuario aprueba el candidato o solicita correcciones. Distinguir «preparado localmente», «aprobado» y «publicado».
5. El usuario gestiona commits, tags, push y publicación. Una recompilación posterior produce un artefacto distinto y debe verificarse antes de presentarlo como el mismo candidato aprobado.

No se requiere un sistema de canales de distribución ni prereleases públicas para este flujo local.

## Datos de desarrollo, pruebas y uso habitual

Acuerdo: perfiles separados de **desarrollo**, **prueba empaquetada** y **uso habitual**, sin cuentas adicionales ni servicios. Son directorios/configuraciones de datos, no tres instalaciones de infraestructura.

- Desarrollo y builds de prueba usan bases, cachés, configuración, logs y salidas de respaldo propias. No reutilizar automáticamente la carpeta sincronizada o la biblioteca habitual.
- Usar datos sintéticos o una copia explícita de un respaldo real cuando se necesite validar el formato. El original no se modifica y la copia sigue siendo privada.
- El perfil de prueba debe ser visible y su configuración no debe heredar silenciosamente rutas de backup de la copia original. Credenciales se configuran por perfil solo si son necesarias.
- Las consultas a proveedores pueden ser reales si se solicita probarlas, sin convertir la prueba en una sincronización de datos personales al proveedor.
- Probar con una copia no incorpora automáticamente sus cambios a la biblioteca habitual. Si el usuario quiere conservar esos cambios, acordar un flujo de exportación/restauración compatible antes de hacerlo.
- Antes de abrir la biblioteca habitual con una versión que migra su esquema, generar copia de seguridad. Volver a un ejecutable anterior no garantiza que pueda leer una base ya migrada; la recuperación debe restaurar también datos compatibles.

El modo portable no implica que todos los perfiles deban escribir junto al EXE: la selección de directorios se concretará al implementar configuración y arranque.

## Convención de versiones acordada

Usar `0.MINOR.PATCH` durante desarrollo inicial: por ejemplo, `0.1.0` para una entrega inicial acordada, `0.2.0` para un conjunto nuevo de funcionalidades y `0.2.1` para una corrección de esa entrega. Son ejemplos, no versiones asignadas.

La versión de la aplicación, la del esquema de base y la del formato de respaldo son conceptos separados. Ningún incremento obliga por sí solo a migrar datos. Documentar compatibilidad y cambios incompatibles expresamente, incluso en versiones `0.x`.

Los resultados locales de pruebas se identifican como builds de prueba sin incrementar automáticamente `MINOR` o `PATCH`. La nomenclatura técnica exacta y la ubicación de artefactos se decidirán al implementar el empaquetado.

## Contenido recomendado de una nota

- Versión y fecha de preparación; fecha de publicación solo si ocurrió.
- Nuevas funcionalidades, cambios y correcciones.
- Formato de datos, migraciones y compatibilidad de respaldos relevantes.
- Requisitos de ejecución y limitaciones conocidas.
- Comprobaciones realizadas y artefactos generados, sin afirmar resultados no ejecutados.

No crear notas vacías para versiones hipotéticas ni una release por cada modificación documental. El alcance y los escenarios de cada comprobación dependen de lo modificado; no ejecutar una batería universal por costumbre.
