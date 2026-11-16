# Estrategia de pruebas

Estado: estrategia aprobada y primeras pruebas implementadas. Pruebas automáticas focalizadas, prioridad a integridad de datos y validación manual de experiencia y paquete.

## Objetivo

Detectar regresiones relevantes, especialmente pérdida de datos y cambios silenciosos de significado, sin convertir cada modificación en una batería extensa. Una prueba comprueba un comportamiento observable: preparar datos conocidos, ejecutar una acción y comparar con un resultado esperado.

## Capas de comprobación

| Capa | Herramienta | Qué comprobar |
| --- | --- | --- |
| Reglas y casos de uso Go | `testing` de Go; dobles pequeños cuando hagan falta | Estados, ciclos, precedencia de datos y decisiones de importación sin arrancar Wails |
| Persistencia e integración local | `testing` con SQLite real en directorios temporales | Guardado, reapertura, relaciones, transacciones y migraciones; no sustituir SQL por un mock que nunca ejecuta las consultas |
| Importadores y adaptadores | `testing`, archivos sintéticos y respuestas HTTP controladas con `httptest` | Mapeo de campos, errores de formato, fechas desconocidas, clasificación y aislamiento; sin depender de APIs reales |
| Interfaz | Vitest y React Testing Library si se adoptan | Interacciones con reglas propias de presentación: selección, revisión, mensajes y confirmación; no una prueba por componente trivial |
| Aplicación empaquetada | Pruebas manuales guiadas inicialmente | Arranque, cierre, persistencia y recorrido afectado dentro de Wails real |

Validación de tipos y compilación complementan las pruebas, pero no demuestran que el comportamiento sea correcto. Un test de React con puente simulado tampoco comprueba la integración real con Wails.

## Ejemplos de pruebas con valor

- Importar una entrada conocida en modo «conservar existentes» deja intactos todos sus datos locales.
- En modo «sobrescribir», el puntaje entrante sustituye al local conforme al mapeo aprobado.
- Repetir el mismo archivo no crea otra copia de los registros ya identificados.
- Una fecha ausente sigue ausente; nunca se transforma en la fecha actual ni de importación.
- La asignación manual en bloque aplica el medio elegido solo a las entradas seleccionadas y no incorpora las ignoradas.
- Refrescar metadata conserva notas y consumo, y no modifica otro medio.
- Guardar y volver a abrir la base recupera los cambios realizados.
- Exportar/resguardar y restaurar en otra ubicación conserva la información incluida en el formato, sin exigir igualdad byte a byte entre archivos SQLite.
- Un fallo durante una operación transaccional no deja una actualización personal a medias; un fallo al preparar backup no sustituye la última copia válida.
- Una migración preserva los datos existentes definidos como compatibles y no reutiliza directorios de uso habitual durante la prueba.

Estos casos se incorporan cuando exista la feature correspondiente; no construir ahora sistemas futuros para poder probarlos.

## Datos, red y aislamiento

- Cada prueba automática crea sus propios archivos/directorios temporales y no consulta rutas de biblioteca o backup del usuario.
- Fixtures públicos pequeños y sintéticos; nunca copiar exportaciones personales reales al repo. Los originales pueden ayudar a explorar el formato localmente, no son dependencia de la suite pública.
- Reloj e IDs controlables cuando el resultado lo necesite; evitar pruebas que dependen de la hora real o de esperas arbitrarias.
- Pruebas de proveedores con respuestas controladas, incluidos errores, sin tokens reales, cuota ni Internet. Las comprobaciones live se ejecutan aparte y explícitamente cuando sean pertinentes.
- Mantener separado el perfil de prueba empaquetada del habitual, incluidas sus salidas de backup. Ver [workflow de releases](releases/README.md).

## Cuándo ejecutar qué

1. **Durante una feature:** comprobar las reglas y fronteras modificadas con la evidencia más directa. Añadir o adaptar tests para lógica relevante; un cambio visual trivial puede verificarse por revisión o interacción sin un nuevo test automático.
2. **Al corregir un bug:** si es una regresión significativa y reproducible, crear primero o junto al arreglo una prueba que fallaría con el comportamiento anterior.
3. **Antes de una build de prueba:** ejecutar comprobaciones relacionadas con los cambios y las dependencias necesarias para producir el paquete. Entregar una lista breve de validación manual.
4. **Al preparar una release:** ejecutar el conjunto estable relevante para lo incluido. Si cambia persistencia, resguardo o migraciones, incluir pruebas de esos recorridos; comprobar también arranque y recorrido principal en el candidato final.

No ejecutar lint, tests, build y automatización visual como ceremonia universal. No repetir suites que ya pasaron si no hubo cambios o dudas nuevas que lo justifiquen. Nunca afirmar que la app completa está probada cuando solo se comprobaron tipos o funciones aisladas.

## Automatización y alcance inicial

Empezar por el backend y los recorridos de persistencia/importación, donde se concentran los riesgos del producto. Automatización end-to-end de escritorio y CI quedan para cuando aporten valor concreto; no se necesita Playwright ni otra herramienta global para salir de planificación.

No fijar un porcentaje de cobertura obligatorio. Priorizar los comportamientos importantes y mantener pocos tests útiles, deterministas y mantenibles, en lugar de pruebas que solo replican la implementación.

Comandos implementados: `go test ./internal/...` para backend, `npm --prefix frontend test` para interfaz y `wails3 task test` para ambos. Vitest y React Testing Library son dependencias de desarrollo locales, no requisitos del usuario final.

Las primeras pruebas cubren validación y fechas desconocidas, persistencia y reapertura SQLite, búsqueda con mayúsculas Unicode, identidad de edición/ciclo, series compartidas, rechazo de esquemas futuros, rollback transaccional y separación de perfiles. En React se comprueba envío del formulario y preservación de datos/mensaje ante rechazo del backend. El puente Wails está simulado en esos tests de React; no prueban por sí mismos el WebView real.

## Reparto de responsabilidades propuesto

El agente implementa y ejecuta las pruebas técnicas pertinentes y comunica qué pasó, qué no se ejecutó y qué queda por revisar. El usuario valida que el comportamiento y la experiencia cumplen su intención mediante modo desarrollo o build de prueba. Esa validación complementa, no sustituye, las comprobaciones técnicas.

El resumen de una entrega incluye cambio realizado, evidencia de verificación y recorridos manuales concretos si los hay. No se crean notas de release para registrar cada corrida de tests.
