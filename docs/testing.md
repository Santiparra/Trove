# Contratos y pruebas

La verificación se organiza por objetivos observables, no por porcentaje de cobertura ni cuota de funciones. El catálogo siguiente contiene **35 escenarios core** actuales. Un escenario puede incluir una admisión y un rechazo pareados; no representa necesariamente un archivo o una función de test.

## Catálogo de aceptación

| ID | Objetivo / admisión y límite | Evidencia automatizada |
| --- | --- | --- |
| C01 | Iniciar biblioteca vacía sin proveedor ni repositorio requerido | `TestAdmission_C01_EmptyDirectoryStartsOffline` |
| C02 | Separar biblioteca y salida de perfiles | `TestAdmission_C02_ProfilesDoNotShareLibraryOrBackupOutput` |
| C03 | Conservar significado de DTOs y versión entre adaptadores | `TestAdmission_C03_WireDTOsPreserveMeaning` + bindings/TypeScript |
| C04 | Rechazar entradas inválidas sin registros parciales | `TestAdmission_C04_InvalidWriteDoesNotCreatePartialRecords` |
| C05 | Recuperar desde SQLite aislado sin JSON/manifiesto | `TestAdmission_C05_SQLiteRecoveryIsStandalone` |
| C06 | Recuperar desde JSON aislado sin SQLite/manifiesto | `TestAdmission_C06_JSONRecoveryIsStandalone` |
| C07 | Distinguir importación externa de restauración propia | `TestAdmission_C07_ExternalFileIsNotARecoveryBackup` |
| C08 | Fallar al configurar respaldo sin bloquear trabajo local | `TestAdmission_C08_BackupConfigurationFailureDoesNotRejectLocalWork` |
| C09 | Rechazar perfil/ruta inválidos sin destino alternativo oculto | `TestAdmission_C09_InvalidProfileOrDataRootHasNoSilentFallback` |
| C10 | Guardar, editar y reabrir conservando identidad/serie | `TestBooksSurviveReopenAndShareSeries` |
| C11 | Revertir transacciones fallidas y conservar edición previa | `TestSaveRollsBackAllTablesOnFailure`, `TestFailedEditPreservesSavedBookAndSeriesCanBeCleared` |
| C12 | Rechazar esquema futuro y migrar el anterior con copia previa | `TestFutureSchemaIsRejected`, `TestVersionOneMigratesAndHasReadableSafetyCopy` |
| C13 | Respetar fechas manuales/desconocidas y cronología | `TestValidationKeepsUnknownDatesAndRejectsInvalidInput`, `TestLiveReadingDates`, `TestSaveAppliesLiveDatesButHonorsManualEdits` |
| C14 | Cambiar estado sin pisar otros datos | `TestQuickStatusChangesOnlyStatusAndUpdatesDashboard` |
| C15 | Ordenar recientes solo por fechas conocidas y limitar preview | `TestDashboardUsesCompletionDatesAndLimitsRecentBooks` |
| C16 | Eliminar un libro sin afectar otro y revertir fallo | `TestDeletePreservesOtherBooksAndRollsBackOnFailure` |
| C17 | Mantener notas, propiedades y orden de serie independientes | `TestPersonalActionsAndSeriesOrdering` |
| C18 | Validar formato, relaciones, campos y contenido faltante | `TestValidationRejectsBrokenRelationshipsAndFutureFormats`, `TestJSONRejectsUnknownFieldsAndTrailingData` |
| C19 | Preservar todas las entidades y fechas en ambos respaldos | `TestArchiveRoundTripPreservesIDsAllCyclesAndUnknownDates` |
| C20 | Conservar dos generaciones válidas y no duplicar sin cambios | `TestAutomaticGenerationsRetentionFailureAndRestoreBothFormats` |
| C21 | No sustituir copia válida por generación incompleta/alterada | `TestIncompleteGenerationsAreIgnoredAndPreparedTamperingIsRejected` |
| C22 | Congelar la preparación y cancelar sin tocar datos activos | `TestInvalidOrModifiedRestoreNeverTouchesActiveData` |
| C23 | Recuperar biblioteca anterior si falla el reemplazo | `TestInstallRollbackLeavesPreviousLibraryReadable` |
| C24 | No importar configuración de otro perfil al restaurar | `TestRestoreKeepsTargetProfileBackupConfiguration` |
| C25 | Resguardar por temporizador, al cierre y tras reinicio | `TestDebounceAndShutdownSavePendingChangesWithoutManualExport` |
| C26 | Esperar escrituras aceptadas y rechazar nuevas al cerrar | `TestShutdownWaitsForAcceptedWritesAndBacksThemUp` |
| C27 | Navegar portal/estados/serie/búsqueda sin guardados implícitos | `frontend/src/App.test.tsx` |
| C28 | Conservar borradores y datos confirmados ante errores de UI | `frontend/src/App.test.tsx` |
| C29 | Exigir confirmación y distinguir éxito/error de recuperación | `frontend/src/backup/BackupPanel.test.tsx` |
| C30 | Mantener el contrato Go/Wails/TypeScript generado | `tools/verify.ps1 -CheckBindings` y compilación TypeScript |
| C31 | Compilar sin Git y consultar identidad desde carpeta vacía | `tools/verify-without-git.ps1` sobre fuentes y portable |
| C32 | Distribuir solo archivos explícitos con versión/perfil/checksums correctos | `tools/build.ps1`, `tools/package.ps1` |
| C33 | Aislar lecturas, series, búsquedas y todas las mutaciones entre Libros y Novelas ligeras | `TestAdmission_C33_ReadingMediaAreIsolatedAcrossAllOperations` + interfaz |
| C34 | Recuperar ambos medios desde JSON o SQLite sin perder identidades | `TestAdmission_C34_MixedLibraryRestoresFromEitherRepresentation` |
| C35 | Admitir respaldos anteriores y migrar sin fabricar fechas ni mezclar bibliotecas | `TestAdmission_C35_LegacyJSONRestoresAsBooksWithoutInventingDates`, `TestSchemaTwoMigrationKeepsBookAndSeriesIdentities`, `TestLegacyDocumentKeepsOriginalContentHash` |

## Ejecución

La entrega de catálogo añade pruebas focalizadas en `internal/catalog`, `internal/providers`, `internal/credentials`, `internal/desktop/catalog_test.go` e `internal/storage/sqlite/catalog_test.go`: validación de ISBN/equivalencias, cascada por edición, protección de credenciales, seguridad de portadas, conservación de datos personales, detección de coincidencias y round-trip de snapshots. También comprueba prioridad de edición exacta, selección inicial informativa y alta con estado/serie sin fechas ficticias. `frontend/src/App.test.tsx` verifica búsqueda flotante local/remota, apertura directa de ficha, reconocimiento local antes del alta, configuración de claves, fallback manual, aislamiento de Novelas ligeras y conservación de borradores ante refrescos exitosos o fallidos. Las pruebas de proveedores usan respuestas sintéticas; no necesitan claves reales ni conectividad.

```powershell
powershell -NoProfile -File tools/verify.ps1
powershell -NoProfile -File tools/build.ps1 -Version development -Profile test -BuildID local
powershell -NoProfile -File tools/verify-without-git.ps1
```

El runner de admisión `tests/contracts` atraviesa los adaptadores que Wails publica y utiliza SQLite real, carpetas nuevas y datos sintéticos. No levanta una ventana, APIs ni nube. La UI se prueba con Vitest/Testing Library y puente controlado; los bindings y sus consumidores se comprueban aparte contra Go.

La compilación sin Git usa fuentes y frontend ya generado en una copia donde Git no descubre el repositorio. El portable consulta identidad con `--version` desde una carpeta vacía, sin abrir una biblioteca. No equivale a automatización gráfica completa.

## Verificación remota

GitHub Actions ejecuta el check **Windows x64** solo en PR hacia `main` o push a `main`. Una suite verde habilita la integración, pero no reemplaza comprobar interacción real, WebView2 y diálogos nativos en una PC. Los artefactos de PR permiten esa validación antes del merge.

La lista crece cuando se añaden capacidades reales. Requisitos ambiguos se resuelven antes de convertirlos en expectativas; casos negativos y recuperación se añaden por el riesgo que protegen, no para aumentar el número de tests.

`frontend/src/books/BookSearch.test.tsx` verifica con reloj controlado la pausa de escritura, la cancelación de temporizadores pendientes y el orden local → proveedor, sin solicitudes reales. Los recorridos de autoría/serie desde fichas locales y remotas y el alta inmediata al elegir estado se verifican en `frontend/src/App.test.tsx`.

`internal/providers/health_test.go` verifica pausa tras fallos, una única recuperación concurrente, solicitudes en cola, invalidación al cambiar credenciales, cancelaciones, latencia sin espera local, 404/429, comprobaciones acotadas, respuestas incompatibles y cierre con cancelación. Usa reloj controlado y servidores HTTP sintéticos, sin claves reales ni servicios externos. `frontend/src/catalog/ProviderHealth.test.tsx` comprueba resumen, estados diferenciados, acción manual y refresco exclusivamente local.
