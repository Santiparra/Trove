# Modelo de dominio en discusión

Este documento describe conceptos, no tablas ni un esquema aprobado.

## Reglas confirmadas

- Cada temporada de anime puede ser una entrada independiente, relacionada con sus secuelas y precuelas.
- Películas y remakes son entradas distintas.
- Cada libro o volumen de novela ligera es una entrada; una serie agrupa esas entradas y permite seguir su progreso.
- Una obra literaria puede tener varias ediciones. La lectura se registra sobre una edición.
- Adaptaciones a otros medios son entradas independientes relacionadas.
- La relectura debe aparecer como estado propio para el usuario.

## Conceptos propuestos

| Concepto | Responsabilidad |
| --- | --- |
| Obra / ítem | Identidad local y datos básicos de lo que se consume |
| Serie / agrupación | Conjunto ordenado de ítems relacionados; no confundir con medio «serie de TV» |
| Edición | Manifestación de una obra literaria: idioma, editorial, formato, etc. |
| Registro personal | Pertenencia a biblioteca, puntaje, notas y estado |
| Sesión o ciclo de consumo | Lectura/visualización concreta, con progreso y fechas; propuesta para conservar repeticiones |
| Relación | Secuela, precuela, adaptación, remake u otro vínculo tipado |
| Vínculo externo | Asociación entre identidad local e identificador de un proveedor |
| Snapshot | Copia persistente de metadata externa con procedencia y fecha |

La representación exacta de estos conceptos se definirá con ejemplos antes de diseñar la base.

## Relecturas: decisión de interfaz, modelado pendiente

El usuario quiere «releyendo» como estado distinto, junto con leyendo, pausado, por leer, sin terminar, etc. No se ha cerrado el catálogo de estados.

Propuesta: mantener ese estado visible y, por debajo, conservar cada ciclo de lectura. Así una segunda lectura no borra la primera ni sus fechas. Falta acordar si el puntaje y las notas son generales o también por ciclo, y cómo se representa el final de una relectura.

## Progreso de serie

Se desea seguimiento individual y del conjunto. Falta definir denominador, ítems opcionales, series en publicación y consumo parcial. La agrupación no debería crear dos registros del mismo hecho de consumo.

## Ejemplos a concretar

- Anime con cuatro temporadas, una película y un especial opcional.
- Serie de libros con volúmenes numerados y un relato intermedio.
- Libro leído en español y releído después en otra edición.
- Novela ligera con adaptación a manga y anime.
- Manga/manhwa seguido por capítulos, sin volúmenes completos registrados.
