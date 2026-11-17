import { useEffect, useRef, useState, type KeyboardEvent } from "react";
import * as Backups from "../../bindings/github.com/Santiparra/Trove/internal/desktop/backups";
import type {
  Preview,
  RestoreResult,
  Status,
} from "../../bindings/github.com/Santiparra/Trove/internal/backup/models";
import Icon from "../components/Icon";
import { errorMessage } from "../books/presentation";

export function phaseLabel(phase: string) {
  return (
    (
      {
        unconfigured: "Configurar respaldo",
        pending: "Respaldo pendiente",
        saving: "Respaldando…",
        current: "Respaldo al día",
        error: "Error de respaldo",
      } as Record<string, string>
    )[phase] || "Respaldo"
  );
}

function timestamp(value: string) {
  if (!value) return "Sin generación comprobada en esta sesión";
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString("es");
}

export default function BackupPanel({
  state,
  onState,
  onClose,
  onRestored,
}: {
  state: Status | null;
  onState: (value: Status) => void;
  onClose: () => void;
  onRestored: () => void;
}) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [preview, setPreview] = useState<Preview | null>(null);
  const [approved, setApproved] = useState(false);
  const [result, setResult] = useState<RestoreResult | null>(null);
  const closeButton = useRef<HTMLButtonElement>(null);
  const panel = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const previous = document.activeElement as HTMLElement | null;
    closeButton.current?.focus();
    return () => previous?.focus();
  }, []);
  async function close() {
    if (busy) return;
    setBusy(true);
    try {
      if (preview) await Backups.CancelRestore(preview.token);
      onClose();
    } catch (reason) {
      setError(errorMessage(reason));
      setBusy(false);
    }
  }
  function keys(event: KeyboardEvent) {
    if (event.key === "Escape") {
      event.preventDefault();
      void close();
    }
    if (event.key === "Tab") {
      const items = panel.current?.querySelectorAll<HTMLElement>(
        "button:not(:disabled),input:not(:disabled)",
      );
      if (!items?.length) {
        event.preventDefault();
        return;
      }
      const first = items[0],
        last = items[items.length - 1];
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault();
        last.focus();
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault();
        first.focus();
      }
    }
  }
  async function action(operation: () => PromiseLike<unknown>) {
    if (busy) return;
    setBusy(true);
    setError("");
    try {
      await operation();
    } catch (reason) {
      setError(errorMessage(reason));
    } finally {
      setBusy(false);
    }
  }
  async function confirm() {
    if (!preview || !approved) return;
    let restored: RestoreResult;
    try {
      restored = await Backups.ConfirmRestore(preview.token);
    } catch (reason) {
      setPreview(null);
      setApproved(false);
      throw reason;
    }
    setResult(restored);
    setPreview(null);
    setApproved(false);
    onRestored();
    // A status refresh failure must not report an already completed restore as failed.
    void Backups.State()
      .then(onState)
      .catch(() => undefined);
  }
  return (
    <div className="dialog-backdrop backup-backdrop">
      <div
        className="backup-panel"
        ref={panel}
        role="dialog"
        aria-modal="true"
        aria-labelledby="backup-title"
        aria-busy={busy}
        onKeyDown={keys}
      >
        <header className="backup-heading">
          <div>
            <p className="eyebrow">
              DATOS LOCALES · PERFIL {state?.profile || "ACTUAL"}
            </p>
            <h2 id="backup-title">Respaldo y restauración</h2>
          </div>
          <button
            ref={closeButton}
            className="icon-button"
            aria-label="Cerrar respaldo"
            disabled={busy}
            onClick={() => void close()}
          >
            <Icon name="close" />
          </button>
        </header>
        <section
          className={`backup-state-card backup-${state?.phase || "unconfigured"}`}
          aria-label="Estado del respaldo"
        >
          <strong>{phaseLabel(state?.phase || "unconfigured")}</strong>
          <span>{timestamp(state?.lastSavedAt || "")}</span>
          {state?.lastError && <p role="alert">{state.lastError}</p>}
        </section>
        <section className="backup-section">
          <h3>Resguardo automático</h3>
          <p>
            Elegí una carpeta una vez. Trove guardará SQLite y JSON después de
            los cambios y al cerrar normalmente, manteniendo la generación
            actual y una anterior.
          </p>
          <code className="backup-path">
            {state?.folder || "Sin carpeta configurada"}
          </code>
          <p className="backup-help">
            Se añade la subcarpeta del perfil para separar pruebas y uso
            habitual. Drive/Dropbox sincronizan estos archivos por su cuenta;
            este estado solo confirma el resguardo local.
          </p>
          <div className="backup-actions">
            <button
              className="secondary"
              disabled={busy}
              onClick={() =>
                void action(async () => onState(await Backups.ChooseFolder()))
              }
            >
              <Icon name="folder" size={16} />
              {state?.folder ? "Cambiar carpeta" : "Elegir carpeta"}
            </button>
            <button
              className="quiet"
              disabled={busy || !state?.folder}
              onClick={() =>
                void action(async () => onState(await Backups.BackupNow()))
              }
            >
              <Icon name="backup" size={16} />
              Respaldar ahora / reintentar
            </button>
          </div>
        </section>
        <section className="backup-section">
          <h3>Recuperar una biblioteca</h3>
          <p>
            Seleccioná un archivo <strong>trove.sqlite</strong> o{" "}
            <strong>trove.json</strong> generado por Trove. Se validará y
            preparará una copia antes de pedirte confirmación. No es una
            importación de StoryGraph, MAL u otros servicios.
          </p>
          <button
            className="secondary"
            disabled={busy}
            onClick={() =>
              void action(async () => {
                try {
                  const value = await Backups.ChooseRestore();
                  if (value) {
                    setPreview(value);
                    setApproved(false);
                    setResult(null);
                  }
                } catch (reason) {
                  setPreview(null);
                  setApproved(false);
                  throw reason;
                }
              })
            }
          >
            <Icon name="refresh" size={16} />
            Seleccionar respaldo para restaurar
          </button>
          {preview && (
            <div className="restore-preview">
              <h4>Biblioteca validada para restaurar</h4>
              <dl>
                <dt>Archivo</dt>
                <dd>{preview.kind}</dd>
                <dt>Generado</dt>
                <dd>{timestamp(preview.createdAt)}</dd>
                <dt>Contenido</dt>
                <dd>
                  {preview.works} obras · {preview.editions} ediciones ·{" "}
                  {preview.cycles} ciclos
                </dd>
                <dt>Destino</dt>
                <dd>Perfil {preview.profile}</dd>
              </dl>
              <code className="backup-path">{preview.source}</code>
              <p className="restore-warning">
                Se reemplazará la biblioteca completa de este perfil, no se
                mezclarán registros. Cerrá otras instancias de Trove. Se
                conservará una copia local de la biblioteca anterior.
              </p>
              {preview.works === 0 && (
                <p className="restore-warning">
                  Esta copia contiene una biblioteca vacía.
                </p>
              )}
              <label className="restore-approval">
                <input
                  type="checkbox"
                  checked={approved}
                  disabled={busy}
                  onChange={(event) => setApproved(event.target.checked)}
                />
                Entiendo que se reemplazará la biblioteca del perfil{" "}
                {preview.profile}.
              </label>
              <div className="backup-actions">
                <button
                  className="quiet"
                  disabled={busy}
                  onClick={() =>
                    void action(async () => {
                      await Backups.CancelRestore(preview.token);
                      setPreview(null);
                      setApproved(false);
                    })
                  }
                >
                  Cancelar preparación
                </button>
                <button
                  className="danger-button"
                  disabled={busy || !approved}
                  onClick={() => void action(confirm)}
                >
                  Restaurar y reemplazar biblioteca
                </button>
              </div>
            </div>
          )}
          {result && (
            <div className="restore-success" role="status">
              <strong>Biblioteca restaurada: {result.works} obras.</strong>
              <p>Copia de la biblioteca anterior:</p>
              <code className="backup-path">{result.safetyCopy}</code>
            </div>
          )}
        </section>
        {error && error !== state?.lastError && (
          <div className="error-banner" role="alert">
            {error}
          </div>
        )}
        {busy && (
          <p className="backup-working" role="status">
            Procesando… Esperá a que termine la operación.
          </p>
        )}
        <p className="backup-help">
          Los respaldos son legibles, sin cifrado propio. Se excluyen portadas,
          cachés, logs y credenciales. El JSON conserva identidades y relaciones
          sin depender del motor SQLite.
        </p>
      </div>
    </div>
  );
}
