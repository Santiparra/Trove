# Uso y datos locales

## Bibliotecas de Libros y Novelas ligeras

Cada sección abre su portal con estados, lecturas actuales, últimas finalizaciones conocidas y distribución visual. Comparten recorridos, pero sus listados, búsquedas y series son independientes. Las vistas previas se limitan al espacio disponible; los listados completos siguen accesibles.

Cada volumen de novela ligera se registra individualmente y puede vincularse a una serie y número de orden. Una serie con el mismo nombre en Libros y Novelas ligeras no se fusiona automáticamente. Cambiar de pestaña no reclasifica ni modifica registros existentes.

- Buscar por título muestra sugerencias locales; Enter o «Ver todos los resultados» abre el listado completo.
- Elegir un estado abre sus registros. Los botones superiores cambian el filtro.
- La serie aparece sobre el título y abre sus libros ordenados por posición; los que no tienen posición quedan al final.
- La ficha permite guardar notas y modificar favorito, propiedad y estado sin abrir el formulario completo.
- Solo el título es obligatorio al crear un libro. ISBN, autoría y fechas pueden completarse después.
- En Novelas ligeras se usa «Añadir volumen», con la misma edición mínima y datos personales. La ficha permite editar, marcar favorito/propiedad, guardar notas y cambiar estados, incluida A continuación.
- Eliminar exige confirmación y elimina los datos asociados al libro, no otros libros de su serie.

## Estados y fechas

Estados: Por leer, A continuación, Leyendo, Leído, Pausado y Sin terminar. A continuación se asigna manualmente.

Al entrar en Leyendo se registra inicio si falta. Al entrar en Leído desde otro estado se registra finalización si hay inicio y falta finalización. Pausar y retomar conserva las fechas existentes. Marcar un libro antiguo como leído sin inicio no genera una finalización ficticia.

Las fechas son editables. La finalización no puede preceder al inicio; pueden coincidir el mismo día. Guardar explícitamente una fecha o borrarla prevalece sobre los automatismos de esa edición. Restaurar no aplica esos automatismos.

Los últimos leídos requieren fecha de finalización para establecer un orden cronológico. Los registros sin ella permanecen en el listado de leídos.

## Perfiles

| Perfil | Ubicación predeterminada |
| --- | --- |
| `habitual` | `data/habitual/` junto al ejecutable de usuario |
| `test` | `data/test/` junto al ejecutable de prueba |
| `development` | `.local/development/` bajo el directorio de ejecución |

El indicador superior muestra perfil, ruta, versión y revisión. También se puede consultar la identidad sin abrir datos:

```powershell
.\Trove.exe --version
```

Una raíz alternativa siempre recibe la subcarpeta del perfil:

```powershell
.\Trove.exe --profile test --data-dir "C:\TrovePruebas"
```

El ejemplo usa `C:\TrovePruebas\test\`. No compartir el mismo directorio activo entre instancias durante una restauración.

## Respaldo automático

1. Abrir **Respaldo y restauración**.
2. Elegir una carpeta distinta de los datos activos. Se agrega el nombre del perfil.
3. Se genera una primera copia. Las modificaciones posteriores se agrupan durante dos segundos y se respaldan automáticamente.

Se conservan dos generaciones completas, cada una con `trove.sqlite`, `trove.json` y `complete.json`. Si no cambió el contenido no se publica una generación duplicada. El cierre normal espera los guardados aceptados y completa pendientes. Un cierre forzado puede dejar cambios solo en la base activa; al reiniciar se comprueba su contenido.

Un fallo de salida conserva el cambio local y la última copia válida. Se informa el error y se reintenta. «Al día» confirma los archivos locales, no el estado de un sincronizador externo. Los respaldos no incluyen portadas, cachés, logs, credenciales ni rutas de configuración.

## Restaurar

Elegir un SQLite o JSON generado por esta función, revisar fecha/contenido/perfil y confirmar el reemplazo completo. Cualquiera de las representaciones válidas funciona sin la otra ni el manifiesto. No es una importación desde servicios externos.

La preparación se valida antes del reemplazo y vence a los diez minutos. Se conserva una copia `before-restore-*.sqlite` en el perfil. Si falla el reemplazo se intenta reabrir la biblioteca anterior. La configuración de salida permanece ligada al perfil de destino.

Si una base dañada impide iniciar, abrir una raíz nueva y restaurar allí sin modificar la anterior:

```powershell
.\Trove.exe --profile test --data-dir "C:\TroveRecuperacion"
```

Los formatos compatibles, límites y relaciones están en [formato de respaldo](backup-format.md).

El esquema local actual es 3. Al abrir datos de las versiones previas se deja una copia `before-schema-3-*.sqlite` antes de incorporar el medio; los registros y series existentes siguen siendo Libros con las mismas identidades. Los respaldos nuevos contienen ambos medios; las copias de formato 1 siguen siendo recuperables como Libros. Restaurar siempre reemplaza el perfil completo, no mezcla su contenido con novelas ligeras posteriores.
