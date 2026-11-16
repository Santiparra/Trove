# Modelo de dominio en discusión

Este documento describe conceptos, no tablas ni un esquema aprobado.

## Reglas confirmadas

- Cada temporada de anime puede ser una entrada independiente, relacionada con sus secuelas y precuelas.
- Películas y remakes son entradas distintas.
- Cada libro o volumen de novela ligera es una entrada; una serie agrupa esas entradas y permite seguir su progreso.
- Una obra literaria puede tener varias ediciones. La lectura se registra sobre una edición.
- Adaptaciones a otros medios son entradas independientes relacionadas.
- La relectura debe aparecer como estado propio para el usuario.
- Cada lectura/relectura y visualización/revisualización es un ciclo independiente, sin borrar los anteriores.
- Se pueden crear obras y ediciones manualmente sin metadata externa y vincularlas posteriormente si aparece una coincidencia. La ausencia del proveedor no bloquea su uso.
- El idioma español de la interfaz no determina el idioma del título. Se prevén títulos en inglés y título original como información complementaria; campos y prioridades precisos pendientes.

## Conceptos propuestos

| Concepto | Responsabilidad |
| --- | --- |
| Obra / ítem | Identidad local y datos básicos de lo que se consume |
| Serie / agrupación | Conjunto ordenado de ítems relacionados; no confundir con medio «serie de TV» |
| Edición | Manifestación de una obra literaria: idioma, editorial, formato, etc. |
| Registro personal | Pertenencia a biblioteca, puntaje, notas y estado |
| Sesión o ciclo de consumo | Lectura/visualización concreta e independiente; progreso y fechas asociados según alcance por definir |
| Relación | Secuela, precuela, adaptación, remake u otro vínculo tipado |
| Vínculo externo | Asociación entre identidad local e identificador de un proveedor |
| Snapshot | Copia persistente de metadata externa con procedencia y fecha |

La representación exacta de estos conceptos se definirá con ejemplos antes de diseñar la base.

## Relecturas: ciclos confirmados, detalles pendientes

El usuario quiere «releyendo» como estado distinto, junto con leyendo, pausado, por leer, sin terminar, etc. No se ha cerrado el catálogo de estados.

Acuerdo: mantener ese estado visible y, por debajo, conservar cada ciclo de lectura. Así una segunda lectura no borra la primera ni sus fechas. Falta acordar si el puntaje y las notas son generales o también por ciclo, y cómo se representa el final de una relectura.

## Progreso de serie

Se desea seguimiento individual y del conjunto. Falta definir denominador, ítems opcionales, series en publicación y consumo parcial. La agrupación no debería crear dos registros del mismo hecho de consumo.

## Registro personal y fechas

- **Favorito:** indicador personal activable/desactivable, independiente del puntaje y del estado de lectura. Alimentará la selección de favoritos de todos los tiempos.
- **En propiedad:** indicador activable/desactivable de la edición local. La interfaz actual trabaja con una edición por libro; gestión multiedición queda para después.
- **A continuación** (`up-next`): estado manual para próximas lecturas, especialmente continuaciones de una serie en curso. «Por leer» se reserva en el uso previsto a nuevas series y obras independientes, mientras «Leyendo» representa consumo actual. No se agrega ni se cambia automáticamente el estado de otros libros de una serie; no se exige una serie local completa para usarlo.

### Fechas al cambiar estados

Reglas implementadas en Go para acciones locales (selector rápido y guardado de formulario):

1. Al marcar una lectura como «Leyendo» por primera vez, registrar la fecha local actual si no tiene inicio. Pausar y retomar no cambia un inicio existente.
2. Al entrar en «Leído» desde cualquier otro estado, registrar finalización actual si hay fecha de inicio y todavía no hay fecha de finalización. No reemplazar fechas ya conocidas. Guardar otra vez el mismo estado no regenera fechas.
3. Marcar directamente como «Leído» un libro sin inicio no genera una fecha de finalización. Permite rescatar consumo antiguo sin atribuirlo al día del registro.
4. Las fechas siguen siendo editables manualmente y pertenecen al ciclo, no a todos los consumos de la obra. La edición explícita debe prevalecer sobre automatismos.
5. Importar un estado antiguo no es una transición de lectura en vivo: las fechas desconocidas de archivos siguen desconocidas. No aplicar estos automatismos a importaciones o restauraciones.

Se permite empezar y terminar el mismo día, nunca una finalización anterior al inicio. La transición automática también se valida: un inicio futuro no puede generar una finalización de hoy que lo preceda. El formulario identifica qué fechas se editaron explícitamente, incluido borrarlas, para respetar esa elección al guardar. Estos automatismos no reconstruyen fechas en datos existentes al abrir/migrar la base.

## Ejemplos a concretar

- Anime con cuatro temporadas, una película y un especial opcional.
- Serie de libros con volúmenes numerados y un relato intermedio.
- Libro leído en español y releído después en otra edición.
- Novela ligera con adaptación a manga y anime.
- Manga/manhwa seguido por capítulos, sin volúmenes completos registrados.
